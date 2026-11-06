import * as BetterQueue from 'better-queue';
import { EventEmitter } from 'events';
import { IQueue } from '../../interfaces/queue.interface';
import { IJob } from '../../interfaces/job.interface';
import { JobOptions, BulkJobOptions, QueueStats } from '../../interfaces/queue-options.interface';
import { JobStatus } from '../../enums/job-status.enum';
import { QueueEvent } from '../../enums/queue-event.enum';
import { BetterQueueJobAdapter } from './better-queue-job.adapter';

interface BetterQueueTask<T> {
  id: string;
  name: string;
  data: T;
  options?: JobOptions;
}

let storeUuid = 0;

function stableSort<T>(arr: T[], compare: (a: T, b: T) => number): T[] {
  const wrapper = arr.map((item, idx) => ({ item, idx }));
  wrapper.sort((a, b) => compare(a.item, b.item) || a.idx - b.idx);
  return wrapper.map((w) => w.item);
}

/**
 * Simple in-memory store for better-queue to avoid dynamic require issues
 */
class MemoryStore {
  private queue: string[] = [];
  private tasks: Record<string, any> = {};
  private priorities: Record<string, number> = {};
  private running: Record<string, Record<string, any>> = {};

  connect(cb: (error: Error | null, length?: number) => void): void {
    cb(null, this.queue.length);
  }

  getTask(taskId: string, cb: (error: Error | null, task?: any) => void): void {
    cb(null, this.tasks[taskId]);
  }

  deleteTask(taskId: string, cb: (error?: Error) => void): void {
    const hadTask = this.tasks[taskId];
    delete this.tasks[taskId];
    delete this.priorities[taskId];
    if (hadTask) {
      const idx = this.queue.indexOf(taskId);
      if (idx !== -1) this.queue.splice(idx, 1);
    }
    cb();
  }

  putTask(taskId: string, task: any, priority: number, cb: (error?: Error) => void): void {
    const hadTask = this.tasks[taskId];
    this.tasks[taskId] = task;
    if (!hadTask) {
      this.queue.push(taskId);
    }
    if (priority !== undefined) {
      this.priorities[taskId] = priority;
      this.queue = stableSort(this.queue, (a, b) => {
        if (this.priorities[a] < this.priorities[b]) return 1;
        if (this.priorities[a] > this.priorities[b]) return -1;
        return 0;
      });
    }
    cb();
  }

  takeFirstN(n: number, cb: (error: Error | null, lockId?: string) => void): void {
    const lockId = String(storeUuid++);
    const taskIds = this.queue.splice(0, n);
    const tasks: Record<string, any> = {};
    taskIds.forEach((taskId) => {
      tasks[taskId] = this.tasks[taskId];
      delete this.tasks[taskId];
    });
    if (taskIds.length > 0) {
      this.running[lockId] = tasks;
    }
    cb(null, lockId);
  }

  takeLastN(n: number, cb: (error: Error | null, lockId?: string) => void): void {
    const lockId = String(storeUuid++);
    const taskIds = this.queue.splice(-n).reverse();
    const tasks: Record<string, any> = {};
    taskIds.forEach((taskId) => {
      tasks[taskId] = this.tasks[taskId];
      delete this.tasks[taskId];
    });
    if (taskIds.length > 0) {
      this.running[lockId] = tasks;
    }
    cb(null, lockId);
  }

  getLock(lockId: string, cb: (error: Error | null, tasks?: Record<string, any>) => void): void {
    cb(null, this.running[lockId]);
  }

  getRunningTasks(cb: (error: Error | null, running?: Record<string, Record<string, any>>) => void): void {
    cb(null, this.running);
  }

  releaseLock(lockId: string, cb: (error?: Error) => void): void {
    delete this.running[lockId];
    cb();
  }
}

/**
 * Better-Queue Queue Adapter
 */
export class BetterQueueQueueAdapter<T = any> implements IQueue<T> {
  private readonly queue: BetterQueue;
  private readonly eventEmitter = new EventEmitter();
  private readonly jobs = new Map<string, BetterQueueJobAdapter<T>>();
  private readonly completedJobs: BetterQueueJobAdapter<T>[] = [];
  private readonly failedJobs: BetterQueueJobAdapter<T>[] = [];
  private jobIdCounter = 1;
  private isPausedFlag = false;
  private processor?: (job: IJob<T>) => Promise<any>;

  constructor(
    public readonly name: string,
    private readonly config?: any,
  ) {
    // Configure better-queue with custom in-memory store
    // This avoids dynamic require issues with webpack bundling
    const queueConfig = {
      store: new MemoryStore() as any, // Cast to any to satisfy better-queue's Store type
      maxRetries: config?.maxRetries ?? 3,
      retryDelay: config?.retryDelay ?? 1000,
      concurrent: config?.concurrent ?? 5,
      maxTimeout: config?.maxTimeout ?? 60000,
      // Spread additional config but remove store if it exists
      ...(config
        ? Object.fromEntries(
            Object.entries(config as Record<string, unknown>).filter(([key]) => key !== 'store'),
          )
        : {}),
    };

    // Create better-queue instance
    this.queue = new BetterQueue(
      (task: BetterQueueTask<T>, cb: (error: Error | null, result?: any) => void) => {
        void this.processTask(task, cb);
      },
      queueConfig,
    );

    // Set up event forwarding
    this.setupEventListeners();
  }

  private setupEventListeners(): void {
    this.queue.on('task_finish', (taskId: string, result: unknown) => {
      const job = this.jobs.get(taskId) as BetterQueueJobAdapter<T>;
      if (job) {
        job._markCompleted(result);
        this.completedJobs.push(job);
        this.eventEmitter.emit(QueueEvent.JOB_COMPLETED, job);
      }
    });

    this.queue.on('task_failed', (taskId: string, errorMessage: any) => {
      const job = this.jobs.get(taskId) as BetterQueueJobAdapter<T>;
      if (job) {
        const error = errorMessage instanceof Error ? errorMessage : new Error(String(errorMessage));
        job._markFailed(error);
        this.failedJobs.push(job);
        this.eventEmitter.emit(QueueEvent.JOB_FAILED, job, error);
      }
    });

    this.queue.on('task_started', (taskId: string) => {
      const job = this.jobs.get(taskId) as BetterQueueJobAdapter<T>;
      if (job) {
        job._markProcessing();
        this.eventEmitter.emit(QueueEvent.JOB_STARTED, job);
      }
    });
  }

  private async processTask(
    task: BetterQueueTask<T>,
    callback: (error: Error | null, result?: any) => void,
  ): Promise<void> {
    try {
      const job = this.jobs.get(task.id);
      if (!job) {
        return callback(new Error('Job not found'));
      }

      if (this.processor) {
        const result = await this.processor(job);
        callback(null, result);
      } else {
        // Default behavior - just complete the job
        callback(null, { success: true });
      }
    } catch (error) {
      callback(error as Error);
    }
  }

  // ===========================
  // Job Operations
  // ===========================

  add(name: string, data: T, options?: JobOptions): Promise<IJob<T>> {
    const jobId = `job-${this.jobIdCounter++}`;
    const timestamp = Date.now();

    const job = new BetterQueueJobAdapter<T>(jobId, name, data, timestamp);
    this.jobs.set(jobId, job);

    const task: BetterQueueTask<T> = {
      id: jobId,
      name,
      data,
      options,
    };

    // Add to better-queue
    // Note: Priority handling would require configuring a priority function in queue options
    this.queue.push(task);

    this.eventEmitter.emit(QueueEvent.JOB_ADDED, job);
    return Promise.resolve(job);
  }

  async addBulk(jobs: BulkJobOptions<T>[]): Promise<IJob<T>[]> {
    const addedJobs: IJob<T>[] = [];
    for (const { name, data, opts } of jobs) {
      const job = await this.add(name, data, opts);
      addedJobs.push(job);
    }
    return addedJobs;
  }

  // ===========================
  // Job Retrieval
  // ===========================

  getJob(jobId: string): Promise<IJob<T> | null> {
    return Promise.resolve(this.jobs.get(jobId) || null);
  }

  async getJobs(statuses: JobStatus[]): Promise<IJob<T>[]> {
    const allJobs = Array.from(this.jobs.values());
    const filteredJobs: IJob<T>[] = [];

    for (const job of allJobs) {
      const state = await job.getState();
      if (statuses.includes(state)) {
        filteredJobs.push(job);
      }
    }

    return filteredJobs;
  }

  async getWaiting(): Promise<IJob<T>[]> {
    return this.getJobs([JobStatus.WAITING]);
  }

  async getActive(): Promise<IJob<T>[]> {
    return this.getJobs([JobStatus.ACTIVE]);
  }

  getCompleted(): Promise<IJob<T>[]> {
    return Promise.resolve(this.completedJobs);
  }

  getFailed(): Promise<IJob<T>[]> {
    return Promise.resolve(this.failedJobs);
  }

  // ===========================
  // Queue Control
  // ===========================

  pause(): Promise<void> {
    this.queue.pause();
    this.isPausedFlag = true;
    this.eventEmitter.emit(QueueEvent.QUEUE_PAUSED);
    return Promise.resolve();
  }

  resume(): Promise<void> {
    this.queue.resume();
    this.isPausedFlag = false;
    this.eventEmitter.emit(QueueEvent.QUEUE_RESUMED);
    return Promise.resolve();
  }

  async close(): Promise<void> {
    await new Promise<void>((resolve) => {
      this.queue.destroy(() => resolve());
    });
    this.jobs.clear();
    this.completedJobs.length = 0;
    this.failedJobs.length = 0;
  }

  isPaused(): Promise<boolean> {
    return Promise.resolve(this.isPausedFlag);
  }

  // ===========================
  // Maintenance
  // ===========================

  clean(grace: number, limit: number, status: JobStatus): Promise<void> {
    if (status === JobStatus.COMPLETED) {
      const toRemove = this.completedJobs.slice(0, limit);
      toRemove.forEach((job) => this.jobs.delete(job.id));
      this.completedJobs.splice(0, limit);
    } else if (status === JobStatus.FAILED) {
      const toRemove = this.failedJobs.slice(0, limit);
      toRemove.forEach((job) => this.jobs.delete(job.id));
      this.failedJobs.splice(0, limit);
    }
    return Promise.resolve();
  }

  async obliterate(): Promise<void> {
    await this.close();
  }

  drain(): Promise<void> {
    this.jobs.clear();
    this.completedJobs.length = 0;
    this.failedJobs.length = 0;
    return Promise.resolve();
  }

  // ===========================
  // Stats
  // ===========================

  async getStats(): Promise<QueueStats> {
    const allJobs = Array.from(this.jobs.values());
    const stats = {
      waiting: 0,
      active: 0,
      completed: this.completedJobs.length,
      failed: this.failedJobs.length,
      delayed: 0,
      paused: 0,
    };

    for (const job of allJobs) {
      const state = await job.getState();
      if (state === JobStatus.WAITING) stats.waiting++;
      else if (state === JobStatus.ACTIVE) stats.active++;
    }

    return stats;
  }

  // ===========================
  // Event Listeners
  // ===========================

  on(event: QueueEvent, handler: (...args: any[]) => void): void {
    this.eventEmitter.on(event, handler);
  }

  off(event: QueueEvent, handler: (...args: any[]) => void): void {
    this.eventEmitter.off(event, handler);
  }

  once(event: QueueEvent, handler: (...args: any[]) => void): void {
    this.eventEmitter.once(event, handler);
  }

  /**
   * Get the underlying better-queue instance (for advanced use cases)
   */
  getUnderlyingQueue(): BetterQueue {
    return this.queue;
  }

  /**
   * Set processor for this queue
   */
  setProcessor(processor: (job: IJob<T>) => Promise<any>): void {
    this.processor = processor;
  }
}

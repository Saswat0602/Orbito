import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';
import { UsersSeederService } from './users.seeder.service';
import { OrganizationsSeederService } from './organizations.seeder.service';
import { WorkspacesSeederService } from './workspaces.seeder.service';
import { ProjectsSeederService } from './projects.seeder.service';
import { WorkflowSeederService } from './workflow.seeder';
import { TaskStatusSeederService } from './taskstatus.seeder.service';
import { TasksSeederService } from './tasks.seeder.service';
import { SprintsSeederService } from './sprints.seeder.service';
import { LabelsSeederService } from './labels.seeder.service';
import { TaskCommentsSeederService } from './task-comments.seeder.service';
import { TaskDependenciesSeederService } from './task-dependencies.seeder.service';
import { TaskWatchersSeederService } from './task-watchers.seeder.service';
import { TimeEntriesSeederService } from './time-entries.seeder.service';
import { AdminSeederService } from './admin-seeder.service';
import { InboxRulesSeederService } from './inbox-rules.seeder.service';

@Injectable()
export class SeederService {
  constructor(
    private prisma: PrismaService,
    private adminSeeder: AdminSeederService,
    private usersSeeder: UsersSeederService,
    private organizationsSeeder: OrganizationsSeederService,
    private workspacesSeeder: WorkspacesSeederService,
    private workflowsSeeder: WorkflowSeederService,
    private projectsSeeder: ProjectsSeederService,
    private taskStatusSeeder: TaskStatusSeederService,
    private tasksSeeder: TasksSeederService,
    private sprintsSeeder: SprintsSeederService,
    private labelsSeeder: LabelsSeederService,
    private taskCommentsSeeder: TaskCommentsSeederService,
    private taskDependenciesSeeder: TaskDependenciesSeederService,
    private taskWatchersSeeder: TaskWatchersSeederService,
    private timeEntriesSeeder: TimeEntriesSeederService,
    private inboxRulesSeeder: InboxRulesSeederService,
  ) {}

  async seedCoreModules() {
    console.log('🌱 Starting core modules seeding...');

    try {
      // 1. Seed Users (foundation)
      const users = await this.usersSeeder.seed();
      console.log('✅ Users seeded');

      // 2. Seed Organizations (depends on users)
      const organizations = await this.organizationsSeeder.seed(users);
      console.log('✅ Organizations seeded');

      // 2.5 Seed Workflows and Task Statuses (depends on organizations)
      const workflows = await this.workflowsSeeder.seed(organizations);
      console.log('✅ Workflows seeded');
      const taskStatuses = await this.taskStatusSeeder.seed(workflows, users);
      console.log('✅ Task statuses seeded');

      // 3. Seed Workspaces (depends on organizations)
      const workspaces = await this.workspacesSeeder.seed(organizations, users);
      console.log('✅ Workspaces seeded');

      // 4. Seed Projects (depends on workspaces and users)
      const projects = await this.projectsSeeder.seed(workspaces, users);
      console.log('✅ Projects seeded');

      // 4.5 Seed Sprints (depends on projects and users)
      const sprints = await this.sprintsSeeder.seed(projects, users);
      console.log('✅ Sprints seeded');

      // 5. Seed Inbox Rules (depends on projects with inboxes)
      await this.seedInboxRules();

      // 7. Seed Tasks (depends on projects, users, and task statuses)
      const tasks = await this.tasksSeeder.seed(projects, users);
      console.log('✅ Tasks seeded');

      // 8. Seed Labels (depends on projects and users)
      const labels = await this.labelsSeeder.seed(projects, users);
      console.log('✅ Labels seeded');

      // 9. Seed Task Comments (depends on tasks and users)
      const taskComments = await this.taskCommentsSeeder.seed(tasks, users);
      console.log('✅ Task comments seeded');

      // 10. Seed Task Dependencies (depends on tasks and users)
      const taskDependencies = await this.taskDependenciesSeeder.seed(tasks, users);
      console.log('✅ Task dependencies seeded');

      // 11. Seed Task Watchers (depends on tasks and users)
      const taskWatchers = await this.taskWatchersSeeder.seed(tasks, users);
      console.log('✅ Task watchers seeded');

      // 12. Seed Time Entries (depends on tasks and users)
      const timeEntries = await this.timeEntriesSeeder.seed(tasks, users);
      console.log('✅ Time entries seeded');

      console.log('🎉 Core modules seeding completed successfully!');

      return {
        users,
        organizations,
        workspaces,
        projects,
        tasks,
        sprints,
        labels,
        taskComments,
        taskDependencies,
        taskWatchers,
        timeEntries,
      };
    } catch (_error) {
      console.error('❌ Error seeding core modules:', _error);
      throw _error;
    }
  }
  async adminSeedModules() {
    console.log('🌱 Starting admin modules seeding...');

    try {
      // 0. Seed Admin User (must be first)
      const adminUser = await this.adminSeeder.seed();
      console.log('✅ Admin user seeded');
      return {
        adminUser,
      };
    } catch (_error) {
      console.error('❌ Error seeding core modules:', _error);
      throw _error;
    }
  }

  async seedInboxRules() {
    console.log('📧 Starting inbox rules seeding...');

    try {
      const result = await this.inboxRulesSeeder.seedRulesForAllInboxes();
      console.log(
        `✅ Inbox rules seeded: ${result.totalCreated} rules created across ${result.inboxesProcessed} inboxes`,
      );
      return result;
    } catch (_error) {
      console.error('❌ Error seeding inbox rules:', _error);
      throw _error;
    }
  }

  async clearCoreModules() {
    console.log('🧹 Clearing core modules...');

    try {
      // Clear in reverse dependency order to avoid foreign key constraints

      // Clear task-related data first
      await this.timeEntriesSeeder.clear();
      console.log('✅ Time entries cleared');

      await this.taskWatchersSeeder.clear();
      console.log('✅ Task watchers cleared');

      await this.taskDependenciesSeeder.clear();
      console.log('✅ Task dependencies cleared');

      await this.taskCommentsSeeder.clear();
      console.log('✅ Task comments cleared');

      await this.labelsSeeder.clear();
      console.log('✅ Labels cleared');

      await this.sprintsSeeder.clear();
      console.log('✅ Sprints cleared');

      await this.tasksSeeder.clear();
      console.log('✅ Tasks cleared');

      // Clear foundation data
      await this.taskStatusSeeder.clear();
      console.log('✅ Task statuses cleared');

      await this.workflowsSeeder.clear();
      console.log('✅ Workflows cleared');

      await this.projectsSeeder.clear();
      console.log('✅ Projects cleared');

      await this.workspacesSeeder.clear();
      console.log('✅ Workspaces cleared');

      await this.organizationsSeeder.clear();
      console.log('✅ Organizations cleared');

      await this.usersSeeder.clear();
      console.log('✅ Users cleared');

      await this.adminSeeder.clear();

      console.log('✅ Admin user cleared');

      console.log('🎉 Core modules cleared successfully!');
    } catch (_error) {
      console.error('❌ Error clearing core modules:', _error);
      throw _error;
    }
  }
}

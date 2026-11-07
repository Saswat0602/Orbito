import { Injectable, Logger } from '@nestjs/common';
import { PrismaService } from '../../prisma/prisma.service';
import { JiraSyncService } from './jira-sync.service';

@Injectable()
export class JiraSyncScheduler {
  private readonly logger = new Logger(JiraSyncScheduler.name);

  constructor(
    private readonly prisma: PrismaService,
    private readonly jiraSyncService: JiraSyncService,
  ) {}

  /**
   * Periodic Jira sync scheduler disabled per user request.
   */
  async processScheduledSyncs() {
    // Jira sync scheduler stopped
    return;
  }
}

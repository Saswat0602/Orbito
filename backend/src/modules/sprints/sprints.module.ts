import { Module } from '@nestjs/common';
import { SprintsService } from './sprints.service';
import { SprintsController } from './sprints.controller';
import { PrismaModule } from '../../prisma/prisma.module';
import { AccessControlService } from '../../common/access-control.utils';
import { ActivityLogService } from '../activity-log/activity-log.service';
import { GatewayModule } from '../../gateway/gateway.module';

@Module({
  imports: [PrismaModule, GatewayModule],
  controllers: [SprintsController],
  providers: [SprintsService, AccessControlService, ActivityLogService],
  exports: [SprintsService],
})
export class SprintsModule {}

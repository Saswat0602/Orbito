import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication, HttpStatus, ValidationPipe } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from './../src/app.module';
import { PrismaService } from './../src/prisma/prisma.service';
import { JwtService } from '@nestjs/jwt';
import { Role, TriggerType, ActionType, RuleStatus } from '@prisma/client';

describe('AutomationController (e2e)', () => {
  let app: INestApplication;
  let prisma: PrismaService;
  let jwtService: JwtService;

  let testUser: any;
  let token: string;
  let organization: any;
  let workspace: any;
  let project: any;
  let createdRuleId: string;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    app.useGlobalPipes(
      new ValidationPipe({
        whitelist: true,
        transform: true,
        forbidNonWhitelisted: true,
      }),
    );
    await app.init();

    prisma = app.get<PrismaService>(PrismaService);
    jwtService = app.get<JwtService>(JwtService);

    // Create user
    testUser = await prisma.user.create({
      data: {
        email: `auto-test-${Date.now()}@example.com`,
        password: 'Password123!',
        firstName: 'Automation',
        lastName: 'Tester',
        username: `autotest_${Date.now()}`,
        role: Role.SUPER_ADMIN,
      },
    });

    token = jwtService.sign({
      sub: testUser.id,
      email: testUser.email,
      role: testUser.role,
    });

    // Create organization, workspace, project for scoping
    organization = await prisma.organization.create({
      data: {
        name: `Auto Org ${Date.now()}`,
        slug: `auto-org-${Date.now()}`,
        ownerId: testUser.id,
      },
    });

    workspace = await prisma.workspace.create({
      data: {
        name: `Auto Workspace ${Date.now()}`,
        slug: `auto-workspace-${Date.now()}`,
        organizationId: organization.id,
        createdBy: testUser.id,
      },
    });

    // Create Workflow
    const workflow = await prisma.workflow.create({
      data: {
        name: `Auto Workflow ${Date.now()}`,
        organizationId: organization.id,
      },
    });

    project = await prisma.project.create({
      data: {
        name: `Auto Project ${Date.now()}`,
        slug: `auto-proj-${Date.now()}`,
        workspaceId: workspace.id,
        workflowId: workflow.id,
        color: '#000000',
        createdBy: testUser.id,
      },
    });
  });

  afterAll(async () => {
    if (prisma) {
      if (createdRuleId) {
        await prisma.ruleExecution.deleteMany({ where: { ruleId: createdRuleId } }).catch(() => {});
        await prisma.automationRule.deleteMany({ where: { id: createdRuleId } }).catch(() => {});
      }
      if (project) {
        await prisma.project.deleteMany({ where: { id: project.id } }).catch(() => {});
      }
      if (organization) {
        await prisma.workflow.deleteMany({ where: { organizationId: organization.id } }).catch(() => {});
      }
      if (workspace) {
        await prisma.workspace.deleteMany({ where: { id: workspace.id } }).catch(() => {});
      }
      if (organization) {
        await prisma.organization.deleteMany({ where: { id: organization.id } }).catch(() => {});
      }
      if (testUser) {
        await prisma.user.deleteMany({ where: { id: testUser.id } }).catch(() => {});
      }
    }
    await app.close();
  });

  describe('Automation Rules CRUD', () => {
    it('POST /api/automation/rules - should create an automation rule', async () => {
      const createDto = {
        name: 'Auto Assign Critical Bugs',
        description: 'Assign critical bugs automatically to project lead',
        triggerType: TriggerType.TASK_CREATED,
        triggerConfig: { priority: 'URGENT' },
        actionType: ActionType.ASSIGN_TASK,
        actionConfig: { assigneeId: testUser.id },
        organizationId: organization.id,
        workspaceId: workspace.id,
        projectId: project.id,
        createdBy: testUser.id,
        status: RuleStatus.ACTIVE,
      };

      const res = await request(app.getHttpServer())
        .post('/api/automation/rules')
        .set('Authorization', `Bearer ${token}`)
        .send(createDto)
        .expect(HttpStatus.CREATED);

      expect(res.body).toHaveProperty('id');
      expect(res.body.name).toBe(createDto.name);
      expect(res.body.triggerType).toBe(TriggerType.TASK_CREATED);
      expect(res.body.actionType).toBe(ActionType.ASSIGN_TASK);
      expect(res.body.status).toBe(RuleStatus.ACTIVE);
      expect(res.body.projectId).toBe(project.id);

      createdRuleId = res.body.id;
    });

    it('GET /api/automation/rules - should list rules filtered by projectId', async () => {
      const res = await request(app.getHttpServer())
        .get(`/api/automation/rules?projectId=${project.id}`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(Array.isArray(res.body)).toBe(true);
      expect(res.body.length).toBeGreaterThanOrEqual(1);
      const found = res.body.find((r: any) => r.id === createdRuleId);
      expect(found).toBeDefined();
      expect(found.name).toBe('Auto Assign Critical Bugs');
    });

    it('GET /api/automation/rules/:id - should get rule by ID', async () => {
      const res = await request(app.getHttpServer())
        .get(`/api/automation/rules/${createdRuleId}`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(res.body.id).toBe(createdRuleId);
      expect(res.body.name).toBe('Auto Assign Critical Bugs');
      expect(res.body).toHaveProperty('executions');
    });

    it('PATCH /api/automation/rules/:id - should update rule configuration', async () => {
      const updateData = {
        name: 'Updated Critical Bug Rule',
        description: 'Updated description for testing',
      };

      const res = await request(app.getHttpServer())
        .patch(`/api/automation/rules/${createdRuleId}`)
        .set('Authorization', `Bearer ${token}`)
        .send(updateData)
        .expect(HttpStatus.OK);

      expect(res.body.name).toBe(updateData.name);
      expect(res.body.description).toBe(updateData.description);
    });

    it('PATCH /api/automation/rules/:id/toggle - should toggle status between ACTIVE and INACTIVE', async () => {
      // Toggle to INACTIVE
      const toggle1 = await request(app.getHttpServer())
        .patch(`/api/automation/rules/${createdRuleId}/toggle`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(toggle1.body.status).toBe(RuleStatus.INACTIVE);

      // Toggle back to ACTIVE
      const toggle2 = await request(app.getHttpServer())
        .patch(`/api/automation/rules/${createdRuleId}/toggle`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(toggle2.body.status).toBe(RuleStatus.ACTIVE);
    });

    it('GET /api/automation/rules/:id/stats - should get rule execution statistics', async () => {
      // Create a mock execution in database
      await prisma.ruleExecution.create({
        data: {
          ruleId: createdRuleId,
          triggerData: { priority: 'URGENT' },
          actionResult: { success: true },
          success: true,
          executionTime: 45,
          triggeredById: testUser.id,
          createdBy: testUser.id,
        },
      });

      const res = await request(app.getHttpServer())
        .get(`/api/automation/rules/${createdRuleId}/stats`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(res.body).toHaveProperty('ruleId', createdRuleId);
      expect(res.body.totalExecutions).toBeGreaterThanOrEqual(1);
      expect(res.body.successCount).toBeGreaterThanOrEqual(1);
      expect(res.body.successRate).toBe(100);
    });

    it('DELETE /api/automation/rules/:id - should delete rule', async () => {
      await request(app.getHttpServer())
        .delete(`/api/automation/rules/${createdRuleId}`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.NO_CONTENT);

      // Verify 404 on subsequent get
      await request(app.getHttpServer())
        .get(`/api/automation/rules/${createdRuleId}`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.NOT_FOUND);
    });
  });
});

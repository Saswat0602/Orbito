import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication, HttpStatus, ValidationPipe } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from './../src/app.module';
import { PrismaService } from './../src/prisma/prisma.service';
import { JwtService } from '@nestjs/jwt';
import { Role, StatusCategory } from '@prisma/client';

describe('ProjectInboxController (e2e)', () => {
  let app: INestApplication;
  let prisma: PrismaService;
  let jwtService: JwtService;

  let testUser: any;
  let token: string;
  let organization: any;
  let workspace: any;
  let project: any;
  let taskStatus: any;
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
        email: `inbox-test-${Date.now()}@example.com`,
        password: 'Password123!',
        firstName: 'Inbox',
        lastName: 'Tester',
        username: `inboxtest_${Date.now()}`,
        role: Role.SUPER_ADMIN,
      },
    });

    token = jwtService.sign({
      sub: testUser.id,
      email: testUser.email,
      role: testUser.role,
    });

    // Create organization, workspace, workflow, project
    organization = await prisma.organization.create({
      data: {
        name: `Inbox Org ${Date.now()}`,
        slug: `inbox-org-${Date.now()}`,
        ownerId: testUser.id,
      },
    });

    workspace = await prisma.workspace.create({
      data: {
        name: `Inbox Workspace ${Date.now()}`,
        slug: `inbox-workspace-${Date.now()}`,
        organizationId: organization.id,
        createdBy: testUser.id,
      },
    });

    const workflow = await prisma.workflow.create({
      data: {
        name: `Inbox Workflow ${Date.now()}`,
        organizationId: organization.id,
      },
    });

    taskStatus = await prisma.taskStatus.create({
      data: {
        name: 'Inbox To Do',
        category: StatusCategory.TODO,
        workflowId: workflow.id,
        color: '#6366f1',
        position: 1,
      },
    });

    project = await prisma.project.create({
      data: {
        name: `Inbox Project ${Date.now()}`,
        slug: `inbox-proj-${Date.now()}`,
        workspaceId: workspace.id,
        workflowId: workflow.id,
        color: '#000000',
        createdBy: testUser.id,
      },
    });
  });

  afterAll(async () => {
    if (prisma) {
      if (project) {
        await prisma.projectInbox.deleteMany({ where: { projectId: project.id } }).catch(() => {});
        await prisma.project.deleteMany({ where: { id: project.id } }).catch(() => {});
      }
      if (organization) {
        await prisma.taskStatus.deleteMany({ where: { workflow: { organizationId: organization.id } } }).catch(() => {});
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

  describe('Project Inbox Operations', () => {
    it('POST /api/projects/:projectId/inbox - should configure project inbox', async () => {
      const inboxDto = {
        name: 'Support & Inbound Requests',
        description: 'Receives customer email requests and converts to tasks',
        emailAddress: `support-${Date.now()}@acme.example.com`,
        autoCreateTask: true,
        defaultStatusId: taskStatus.id,
        syncInterval: '5',
      };

      const res = await request(app.getHttpServer())
        .post(`/api/projects/${project.id}/inbox`)
        .set('Authorization', `Bearer ${token}`)
        .send(inboxDto);

      if (res.status !== HttpStatus.CREATED) {
        console.error('POST INBOX ERROR:', res.status, res.body);
      }

      expect(res.status).toBe(HttpStatus.CREATED);
      expect(res.body).toHaveProperty('id');
      expect(res.body.name).toBe(inboxDto.name);
      expect(res.body.projectId).toBe(project.id);
      expect(res.body.autoCreateTask).toBe(true);
    });

    it('GET /api/projects/:projectId/inbox - should retrieve inbox configuration', async () => {
      const res = await request(app.getHttpServer())
        .get(`/api/projects/${project.id}/inbox`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(res.body).toHaveProperty('id');
      expect(res.body.name).toBe('Support & Inbound Requests');
      expect(res.body.projectId).toBe(project.id);
    });

    it('PUT /api/projects/:projectId/inbox - should update inbox configuration', async () => {
      const updateDto = {
        name: 'Updated Support Inbound',
        description: 'Updated inbox description',
        defaultStatusId: taskStatus.id,
        syncInterval: '10',
      };

      const res = await request(app.getHttpServer())
        .put(`/api/projects/${project.id}/inbox`)
        .set('Authorization', `Bearer ${token}`)
        .send(updateDto);

      if (res.status !== HttpStatus.OK) {
        console.error('PUT INBOX ERROR:', res.status, res.body);
      }

      expect(res.status).toBe(HttpStatus.OK);
      expect(res.body.name).toBe(updateDto.name);
      expect(res.body.description).toBe(updateDto.description);
    });

    it('POST /api/projects/:projectId/inbox/rules - should create inbox filtering rule', async () => {
      const ruleDto = {
        name: 'Urgent Customer Escalation',
        description: 'Auto-prioritize emails matching urgent keywords',
        priority: 10,
        enabled: true,
        conditions: {
          subject: { contains: 'URGENT' },
        },
        actions: {
          setPriority: 'URGENT',
        },
      };

      const res = await request(app.getHttpServer())
        .post(`/api/projects/${project.id}/inbox/rules`)
        .set('Authorization', `Bearer ${token}`)
        .send(ruleDto)
        .expect(HttpStatus.CREATED);

      expect(res.body).toHaveProperty('id');
      expect(res.body.name).toBe(ruleDto.name);
      expect(res.body.priority).toBe(10);
      createdRuleId = res.body.id;
    });

    it('GET /api/projects/:projectId/inbox/rules - should list inbox rules', async () => {
      const res = await request(app.getHttpServer())
        .get(`/api/projects/${project.id}/inbox/rules`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(Array.isArray(res.body)).toBe(true);
      expect(res.body.length).toBeGreaterThanOrEqual(1);
      const found = res.body.find((r: any) => r.id === createdRuleId);
      expect(found).toBeDefined();
      expect(found.name).toBe('Urgent Customer Escalation');
    });

    it('DELETE /api/projects/:projectId/inbox/rules/:ruleId - should delete inbox rule', async () => {
      const res = await request(app.getHttpServer())
        .delete(`/api/projects/${project.id}/inbox/rules/${createdRuleId}`)
        .set('Authorization', `Bearer ${token}`)
        .expect(HttpStatus.OK);

      expect(res.body).toHaveProperty('message', 'Rule deleted successfully');
    });
  });
});

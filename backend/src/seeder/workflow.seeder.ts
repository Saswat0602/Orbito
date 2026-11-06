import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';
import { Workflow, Organization } from '@prisma/client';

@Injectable()
export class WorkflowSeederService {
  constructor(private prisma: PrismaService) {}

  async seed(organizations: Organization[]) {
    console.log('🌱 Seeding workflows...');

    if (!organizations || organizations.length === 0) {
      throw new Error('Organizations must be seeded before workflows');
    }

    const createdWorkflows: Workflow[] = [];

    // Create workflows for each organization
    for (const organization of organizations) {
      const workflowsData = this.getWorkflowsDataForOrganization(organization);

      for (const workflowData of workflowsData) {
        const existingWorkflow = await this.prisma.workflow.findFirst({
          where: {
            name: workflowData.name,
            organizationId: organization.id,
          },
        });

        if (existingWorkflow) {
          console.log(`   ✓ Workflow already exists: ${existingWorkflow.name} in ${organization.name}`);
          createdWorkflows.push(existingWorkflow);
          continue;
        }

        try {
          // Find organization owner to set as creator
          const orgWithOwner = await this.prisma.organization.findUnique({
            where: { id: organization.id },
            include: { owner: true },
          });
          const creatorUser = orgWithOwner?.owner || {
            id: organization.ownerId,
          };

          const workflow = await this.prisma.workflow.create({
            data: {
              ...workflowData,
              organizationId: organization.id,
              createdBy: creatorUser.id,
              updatedBy: creatorUser.id,
            },
          });

          createdWorkflows.push(workflow);
          console.log(`   ✓ Created workflow: ${workflow.name} in ${organization.name}`);
        } catch (error) {
          console.warn(
            `   ⚠ Failed to create workflow ${workflowData.name}: ${error.message}`,
          );
          const fallback = await this.prisma.workflow.findFirst({
            where: {
              name: workflowData.name,
              organizationId: organization.id,
            },
          });
          if (fallback) {
            createdWorkflows.push(fallback);
          }
        }
      }
    }

    console.log(
      `✅ Workflows seeding completed. Created/Found ${createdWorkflows.length} workflows.`,
    );
    return createdWorkflows;
  }

  private getWorkflowsDataForOrganization(organization: Organization) {
    // Create different workflows based on organization type or provide defaults
    const baseWorkflows = [
      {
        name: 'Software Development Workflow',
        description:
          'Standard workflow for software development projects with code review and testing phases',
        isDefault: true,
      },
      {
        name: 'Design & Creative Workflow',
        description:
          'Workflow optimized for design and creative projects with review and approval stages',
        isDefault: false,
      },
      {
        name: 'Marketing Campaign Workflow',
        description: 'Workflow for marketing campaigns from ideation to publication',
        isDefault: false,
      },
      {
        name: 'Client Project Workflow',
        description: 'Workflow for client projects with client review and approval stages',
        isDefault: false,
      },
      {
        name: 'Support & Maintenance Workflow',
        description: 'Simple workflow for support tickets and maintenance tasks',
        isDefault: false,
      },
    ];

    // You can customize workflows based on organization name/type
    if (
      organization.name.toLowerCase().includes('startup') ||
      organization.name.toLowerCase().includes('tech')
    ) {
      return baseWorkflows.slice(0, 3); // Tech-focused workflows
    } else if (
      organization.name.toLowerCase().includes('agency') ||
      organization.name.toLowerCase().includes('creative')
    ) {
      return [baseWorkflows[1], baseWorkflows[2], baseWorkflows[3]]; // Creative-focused workflows
    }

    // Default: return all workflows for comprehensive organizations
    return baseWorkflows;
  }

  async clear() {
    console.log('🧹 Clearing workflows...');

    try {
      // Delete status transitions first
      const deletedTransitions = await this.prisma.statusTransition.deleteMany();
      console.log(`   ✓ Deleted ${deletedTransitions.count} status transitions`);

      // Delete task statuses
      const deletedStatuses = await this.prisma.taskStatus.deleteMany();
      console.log(`   ✓ Deleted ${deletedStatuses.count} task statuses`);

      // Delete workflows
      const deletedWorkflows = await this.prisma.workflow.deleteMany();
      console.log(`✅ Deleted ${deletedWorkflows.count} workflows`);
    } catch (_error) {
      console.error('❌ Error clearing workflows:', _error);
      throw _error;
    }
  }

  findAll() {
    return this.prisma.workflow.findMany({
      select: {
        id: true,
        name: true,
        description: true,
        isDefault: true,
        organization: {
          select: {
            id: true,
            name: true,
            slug: true,
          },
        },
        _count: {
          select: {
            statuses: true,
            transitions: true,
          },
        },
        createdAt: true,
      },
      orderBy: [{ organization: { name: 'asc' } }, { isDefault: 'desc' }, { name: 'asc' }],
    });
  }
}

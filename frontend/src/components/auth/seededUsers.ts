import { Role } from '@orbito/types';

export interface SeededAccount {
  email: string;
  username: string;
  name: string;
  role: Role;
  status: 'ACTIVE' | 'PENDING';
  bio: string;
  avatar?: string;
  title: string;
  password: string;
}

export const DEFAULT_SEEDED_PASSWORD = 'password123';

export const SEEDED_ACCOUNTS: SeededAccount[] = [
  {
    email: 'admin@orbito.io',
    username: 'admin',
    name: 'Admin User',
    role: Role.SUPER_ADMIN,
    status: 'ACTIVE',
    title: 'System Administrator',
    bio: 'System administrator with full access to all features and settings',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'jane.smith@orbito.io',
    username: 'janesmith',
    name: 'Jane Smith',
    role: Role.MANAGER,
    status: 'ACTIVE',
    title: 'Product Manager',
    bio: 'Product Manager with 8+ years of experience in agile development',
    avatar: 'https://images.unsplash.com/photo-1494790108755-2616b612b606?w=150',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'john.doe@orbito.io',
    username: 'johndoe',
    name: 'John Doe',
    role: Role.MEMBER,
    status: 'ACTIVE',
    title: 'Senior Frontend Developer',
    bio: 'Senior Frontend Developer specializing in React and TypeScript',
    avatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=150',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'mike.wilson@orbito.io',
    username: 'mikewilson',
    name: 'Mike Wilson',
    role: Role.MEMBER,
    status: 'ACTIVE',
    title: 'Backend Developer',
    bio: 'Backend Developer focused on Node.js, NestJS, and PostgreSQL',
    avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'sarah.jones@orbito.io',
    username: 'sarahjones',
    name: 'Sarah Jones',
    role: Role.MEMBER,
    status: 'ACTIVE',
    title: 'UI/UX Designer',
    bio: 'UI/UX Designer passionate about creating intuitive user experiences',
    avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=150',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'alex.brown@orbito.io',
    username: 'alexbrown',
    name: 'Alex Brown',
    role: Role.MEMBER,
    status: 'ACTIVE',
    title: 'DevOps Engineer',
    bio: 'DevOps Engineer specializing in cloud infrastructure and automation',
    avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=150',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'emma.davis@orbito.io',
    username: 'emmadavis',
    name: 'Emma Davis',
    role: Role.VIEWER,
    status: 'ACTIVE',
    title: 'QA Analyst',
    bio: 'QA Analyst ensuring quality across all product features',
    avatar: 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=150',
    password: DEFAULT_SEEDED_PASSWORD,
  },
  {
    email: 'tom.garcia@orbito.io',
    username: 'tomgarcia',
    name: 'Tom Garcia',
    role: Role.MEMBER,
    status: 'PENDING',
    title: 'Junior Developer',
    bio: 'Junior Developer eager to learn and contribute (Pending verification)',
    password: DEFAULT_SEEDED_PASSWORD,
  },
];


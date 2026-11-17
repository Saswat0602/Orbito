import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  const baseUrl = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3001';

  return {
    rules: [
      {
        userAgent: '*',
        allow: [
          '/',
          '/intro',
          '/login',
          '/register',
          '/forgot-password',
          '/terms-of-service',
          '/privacy-policy',
        ],
        disallow: [
          '/api/',
          '/admin/',
          '/settings/',
          '/invite/',
          '/reset-password',
          '/*/*/tasks',
          '/*/*/sprints',
          '/*/*/settings',
        ],
      },
    ],
    sitemap: `${baseUrl}/sitemap.xml`,
  };
}

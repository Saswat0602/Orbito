import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/public',
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Read auth cookies set by backend
  const hasAccessToken = request.cookies.has('access_token');
  const hasRefreshToken = request.cookies.has('refresh_token');
  const isAuthenticated = hasAccessToken || hasRefreshToken;

  const isPublicRoute = PUBLIC_ROUTES.some((route) =>
    pathname.startsWith(route)
  );

  // If unauthenticated user tries to access a protected route
  if (!isAuthenticated && !isPublicRoute) {
    const loginUrl = new URL('/login', request.url);
    if (pathname !== '/') {
      loginUrl.searchParams.set('returnUrl', pathname);
    }
    return NextResponse.redirect(loginUrl);
  }

  // If authenticated user visits login or register, redirect them to root workspace resolver
  if (isAuthenticated && (pathname === '/login' || pathname === '/register')) {
    const rootUrl = new URL('/', request.url);
    return NextResponse.redirect(rootUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (svg, png, jpg, etc.)
     * - /api (backend reverse-proxy routes)
     */
    '/((?!_next/static|_next/image|favicon.ico|api|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};


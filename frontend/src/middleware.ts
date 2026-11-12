import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/public',
  '/invite',
  '/setup',
  '/terms-of-service',
  '/privacy-policy',
  '/intro',
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
  // BUT if there is an error/expired/logout parameter, clear invalid cookies and allow login
  if (isAuthenticated && (pathname === '/login' || pathname === '/register')) {
    const hasError =
      request.nextUrl.searchParams.has('error') ||
      request.nextUrl.searchParams.has('expired') ||
      request.nextUrl.searchParams.has('message');

    if (hasError) {
      const response = NextResponse.next();
      response.cookies.delete('access_token');
      response.cookies.delete('refresh_token');
      return response;
    }

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


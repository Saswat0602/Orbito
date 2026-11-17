// common/request-context.interceptor.ts
import { Injectable, NestInterceptor, ExecutionContext, CallHandler } from '@nestjs/common';
import { Observable } from 'rxjs';
import { RequestContextService } from './request-context.service';
import { randomUUID } from 'crypto';

@Injectable()
export class RequestContextInterceptor implements NestInterceptor {
  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const request = context.switchToHttp().getRequest();
    const response = context.switchToHttp().getResponse();

    const correlationId =
      (request?.headers?.['x-correlation-id'] as string) ||
      (request?.headers?.['x-request-id'] as string) ||
      randomUUID();

    // Attach correlation ID to response headers for distributed tracing
    if (response?.setHeader) {
      response.setHeader('x-correlation-id', correlationId);
    }

    const user = request.user; // Provided by your AuthGuard

    // Build a request context object
    const ctx = {
      correlationId,
      userId: user?.id,
      user,
    };

    return RequestContextService.run(ctx, () => next.handle());
  }
}

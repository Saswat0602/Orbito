import {
  ExceptionFilter,
  Catch,
  ArgumentsHost,
  HttpException,
  HttpStatus,
  Logger,
} from '@nestjs/common';
import { Request, Response } from 'express';
import { Prisma } from '@prisma/client';
import { RequestContextService } from '../request-context.service';

interface ErrorResponsePayload {
  statusCode: number;
  message: string | string[] | object;
  error?: string;
  path: string;
  method: string;
  timestamp: string;
  correlationId?: string;
}

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  private readonly logger = new Logger(GlobalExceptionFilter.name);

  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    let status = HttpStatus.INTERNAL_SERVER_ERROR;
    let message: string | string[] | object = 'Internal server error';
    let errorName = 'InternalServerError';

    const correlationId =
      RequestContextService.getCorrelationId() ||
      (request.headers['x-correlation-id'] as string) ||
      (request.headers['x-request-id'] as string);

    // 1. Standard NestJS HTTP Exceptions
    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const res = exception.getResponse();
      if (typeof res === 'string') {
        message = res;
      } else if (typeof res === 'object' && res !== null) {
        const resObj = res as Record<string, unknown>;
        message = (resObj.message as string | string[]) || exception.message;
        errorName = (resObj.error as string) || exception.name;
      } else {
        message = exception.message;
      }
    }
    // 2. Prisma Database Specific Errors
    else if (exception instanceof Prisma.PrismaClientKnownRequestError) {
      switch (exception.code) {
        case 'P2002': {
          // Unique constraint violation
          status = HttpStatus.CONFLICT;
          errorName = 'Conflict';
          const target = (exception.meta?.target as string[]) || [];
          message = target.length
            ? `Duplicate entry detected on unique field(s): ${target.join(', ')}`
            : 'A record with this unique value already exists.';
          break;
        }
        case 'P2025': {
          // Record not found
          status = HttpStatus.NOT_FOUND;
          errorName = 'NotFound';
          message = (exception.meta?.cause as string) || 'Requested record was not found.';
          break;
        }
        case 'P2003': {
          // Foreign key constraint failed
          status = HttpStatus.BAD_REQUEST;
          errorName = 'ForeignKeyViolation';
          const field = (exception.meta?.field_name as string) || 'foreign key';
          message = `Foreign key reference constraint failed on: ${field}`;
          break;
        }
        default: {
          status = HttpStatus.BAD_REQUEST;
          errorName = 'DatabaseError';
          message = 'A database operation error occurred.';
          break;
        }
      }

      this.logger.warn(
        `Prisma ${exception.code} [${status}] on ${request.method} ${request.url} - CID: ${correlationId || 'N/A'}`
      );
    }
    // 3. Generic / Uncaught Exceptions
    else if (exception instanceof Error) {
      message =
        process.env.NODE_ENV === 'production'
          ? 'An unexpected internal error occurred.'
          : exception.message;
      errorName = exception.name;

      this.logger.error(
        `Unhandled Exception [500] on ${request.method} ${request.url} - CID: ${correlationId || 'N/A'} - ${exception.message}`,
        exception.stack
      );
    } else {
      this.logger.error(
        `Unknown Exception [500] on ${request.method} ${request.url} - CID: ${correlationId || 'N/A'}`
      );
    }

    const payload: ErrorResponsePayload = {
      statusCode: status,
      error: errorName,
      message,
      path: request.url,
      method: request.method,
      timestamp: new Date().toISOString(),
      ...(correlationId ? { correlationId } : {}),
    };

    response.status(status).json(payload);
  }
}

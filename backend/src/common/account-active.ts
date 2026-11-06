import { UnauthorizedException } from '@nestjs/common';
import { UserStatus } from '@prisma/client';

/**
 * The rule every token-authenticated entry point has to apply: a token proves
 * who signed in, never that the account is still allowed to be signed in.
 *
 * Without this check a token issued before an administrator deactivated the
 * account keeps working until it expires, so deactivation does not take effect
 * when it is most likely to matter: offboarding, an incident, a policy action.
 * Password login refuses a non-ACTIVE account already, but that only covers the
 * one path that does not use a token.
 *
 * A missing account is refused for the same reason. `DELETE /users/:id` removes
 * the row outright, and a token naming a row that is gone must not authenticate
 * anything.
 *
 * The rule lives here rather than in one of its callers because there are three
 * of them and they cannot share a service: the WebSocket gateway is unable to
 * import UsersService, since UsersModule already imports GatewayModule and Nest
 * would see a cycle. A pure function has no module graph to get caught in.
 */
export function assertAccountActive(account: { status: UserStatus } | null | undefined): void {
  if (!account || account.status !== UserStatus.ACTIVE) {
    throw new UnauthorizedException('Account is not active');
  }
}

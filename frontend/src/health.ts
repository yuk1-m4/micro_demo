export type Health = { status?: unknown }

export function describeHealth(body: Health): string {
  return body.status === 'ok' ? 'ok' : 'unexpected response'
}

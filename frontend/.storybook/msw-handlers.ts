import { http, HttpResponse } from 'msw'
import { makeProject, makeStack, makeUser, type ReleaseEvent, type StackReleaseDetail } from './fixtures'

export const RELEASES_PATH =
  '/api/v1/organizations/:orgId/projects/:projectName/stacks/:stackId/releases'

/** GET release-detail + one-shot events handlers for the deploy timeline. */
export function releaseHandlers(
  details: Record<string, StackReleaseDetail>,
  events: ReleaseEvent[] = [],
) {
  return [
    http.get(`${RELEASES_PATH}/:releaseId/events`, ({ params }) =>
      HttpResponse.json({
        items: events.filter((e) => !e.release_id || e.release_id === params.releaseId),
      }),
    ),
    http.get(`${RELEASES_PATH}/:releaseId`, ({ params }) => {
      const detail = details[params.releaseId as string]
      return detail
        ? HttpResponse.json(detail)
        : HttpResponse.json({ message: 'release not found' }, { status: 404 })
    }),
  ]
}

// Auth endpoints must always resolve: an unmocked 401 sends the axios client
// through its refresh flow and, on failure, hard-redirects the story iframe
// to /sign-in (src/api/client.ts).
export const baselineHandlers = [
  http.get('/api/v1/users/current', () => HttpResponse.json(makeUser())),
  http.post('/api/v1/auth/refresh', () =>
    HttpResponse.json({ token: 'sb-token', refreshToken: 'sb-refresh' }),
  ),
  http.get('/api/v1/organizations/:orgId/projects', () =>
    HttpResponse.json({ items: [makeProject()], total: 1 }),
  ),
  http.get('/api/v1/organizations/:orgId/stacks', () =>
    HttpResponse.json({
      items: [
        makeStack(),
        makeStack({ id: 's2', name: 'billing-worker' }),
        makeStack({ id: 's3', name: 'docs-site' }),
      ],
      total: 3,
    }),
  ),
]

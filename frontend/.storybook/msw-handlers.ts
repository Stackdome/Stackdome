import { http, HttpResponse } from 'msw'
import { makeProject, makeStack, makeUser } from './fixtures'

// Auth endpoints must always resolve: an unmocked 401 sends the axios client
// through its refresh flow and, on failure, hard-redirects the story iframe
// to /sign-in (src/api/client.ts).
export const baselineHandlers = [
  http.get('/api/v1/users/me', () => HttpResponse.json(makeUser())),
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

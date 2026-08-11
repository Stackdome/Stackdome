import React from 'react'
import type { Preview } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router-dom'
import { addons } from 'storybook/preview-api'
import { mswLoader } from 'msw-storybook-addon/csf3'
import '../src/index.css'
import { makeUser } from './fixtures'
import { baselineHandlers } from './msw-handlers'

// Seed before anything renders: the axios interceptor and org-id helpers read
// these keys synchronously, and a missing authToken sends stories into the
// refresh → /sign-in redirect path.
localStorage.setItem('authToken', 'sb-token')
localStorage.setItem('refreshToken', 'sb-refresh')
localStorage.setItem('currentUser', JSON.stringify(makeUser()))

// Driven off the globals channel rather than a decorator: decorators wrap
// stories only, so the Foundations MDX pages — which document the tokens the
// toggle exists to show — would never receive the class.
function applyTheme({ globals }: { globals?: { theme?: string } }) {
  const theme = globals?.theme
  if (!theme) return
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  root.classList.add(theme)
}

const channel = addons.getChannel()
channel.on('setGlobals', applyTheme)
channel.on('globalsUpdated', applyTheme)

const preview: Preview = {
  loaders: [mswLoader()],
  globalTypes: {
    theme: {
      description: 'Color theme',
      toolbar: {
        title: 'Theme',
        icon: 'mirror',
        items: ['light', 'dark'],
        dynamicTitle: true,
      },
    },
  },
  initialGlobals: {
    theme: 'light',
  },
  decorators: [
    (Story) => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
  parameters: {
    msw: baselineHandlers,
    options: {
      storySort: {
        order: ['Foundations', 'Primitives', 'Branded', 'Features', 'Pages'],
      },
    },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview;

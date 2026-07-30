import React from 'react'
import type { Decorator } from '@storybook/react-vite'
import { ConfirmProvider } from '../src/components/branded/confirm'
import { CurrentUserProvider } from '../src/contexts/current-user-context'
import { StackProvider } from '../src/pages/stacks/contexts/stack-context'

export const withConfirm: Decorator = (Story) => (
  <ConfirmProvider>
    <Story />
  </ConfirmProvider>
)

export const withCurrentUser: Decorator = (Story) => (
  <CurrentUserProvider>
    <Story />
  </CurrentUserProvider>
)

export const withStack: Decorator = (Story) => (
  <StackProvider>
    <Story />
  </StackProvider>
)

export const withHeight = (px: number): Decorator =>
  function HeightDecorator(Story) {
    return (
      <div style={{ height: px }}>
        <Story />
      </div>
    )
  }

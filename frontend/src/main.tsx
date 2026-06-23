import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { getAppConfig } from '@/api/config'

// Warm the GitHub-OAuth gate config before React mounts so the auth screens
// render the button without a per-page round-trip (matters on slow networks).
// Consumers fail closed, so the warm-up's own rejection is swallowed here.
void getAppConfig().catch(() => {})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)

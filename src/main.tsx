import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { AuthProvider, useAuth } from './lib/auth'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'

// eslint-disable-next-line react-refresh/only-export-components
function RootApp() {
  // We need to provide the auth context to the router via context
  const auth = useAuth()
  const router = createRouter({
    routeTree,
    context: { auth },
  })
  return <RouterProvider router={router} />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider>
      <RootApp />
    </AuthProvider>
  </StrictMode>,
)

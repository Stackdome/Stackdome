import { Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from "react-router-dom"
import Login from "@/pages/login"
import Signup from "@/pages/signup"
import Dashboard from "@/pages/dashboard"

// Create router with routes
const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path="/" element={<div className="flex flex-col items-center justify-center min-h-svh">Home Page</div>} />
      <Route path="/login" element={<Login />} />
      <Route path="/sign-up" element={<Signup />} />
      <Route path="/dashboard" element={<Dashboard />} />
    </>
  )
)

function App() {
  return <RouterProvider router={router} />
}

export default App

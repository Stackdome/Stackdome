import { GalleryVerticalEnd } from "lucide-react"
import { Link } from "react-router-dom"
import { SignupForm } from "@/pages/signup/components/signup-form"

export default function Signup() {
  return (
    <div className="container relative flex min-h-svh flex-col items-center justify-center md:grid lg:max-w-none lg:grid-cols-2 lg:px-0">
      <Link 
        to="/login" 
        className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 absolute right-4 top-4 md:right-8 md:top-8"
      >
        Login
      </Link>
      
      <div className="relative hidden h-full flex-col bg-zinc-900 p-10 text-white dark:border-r lg:flex">
        <div className="relative z-20 flex items-center text-lg font-medium">
          <GalleryVerticalEnd className="h-6 w-6" />
          Stackdome
        </div>
        <div className="relative z-20 mt-auto">
          <blockquote className="space-y-2">
            <p className="text-lg">"Cloud infrastructure made effortless"</p>
            <footer className="text-sm">Built with 🤍</footer>
          </blockquote>
        </div>
      </div>
      
      <div className="lg:p-8 flex items-center justify-center">
        <SignupForm />
      </div>
    </div>
  )
}

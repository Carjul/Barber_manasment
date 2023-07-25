"use client"
import { useAppDispatch, useAppSelector } from '../../data/store'
import { UserButton } from "@clerk/nextjs";

export default function Home() {
    const dispatch = useAppDispatch()
    const state = useAppSelector(state => state)

  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-24">
      <div className="z-10 w-full max-w-5xl items-center justify-between font-mono text-sm lg:flex">
       home
       <UserButton afterSignOutUrl="/"/>
      </div>
    </main>
  )
}

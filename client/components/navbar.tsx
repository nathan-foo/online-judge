"use client";

import Link from "next/link";
import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";


const Navbar = () => {
    return (
        <nav className="px-4 py-2 bg-transparent fixed top-0 left-0 right-0 z-50 border-b border-transparent">
            <div className="max-w-5xl mx-auto w-full flex justify-between items-center">
                <Link href="/" className="font-bold text-sm">
                    Online Judge
                </Link>
                <div className="flex items-center justify-center">
                    <Button variant="link">
                        <Link href="/dashboard">Dashboard</Link>
                    </Button>
                </div>
                <Show when="signed-out">
                    <div className="flex gap-2">
                        <SignInButton>
                            <Button>
                                Sign In
                            </Button>
                        </SignInButton>
                        <SignUpButton>
                            <Button variant='outline'>
                                Sign Up
                            </Button>
                        </SignUpButton>
                    </div>
                </Show>
                <Show when="signed-in">
                    <UserButton />
                </Show>
            </div>
        </nav>
    );
}

export default Navbar;
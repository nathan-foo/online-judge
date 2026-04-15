"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";

const SCROLL_THRESHOLD = 50;

const Navbar = () => {
    const [scrolled, setScrolled] = useState(false);

    useEffect(() => {
        const onScroll = () => setScrolled(window.scrollY > SCROLL_THRESHOLD);
        onScroll();
        window.addEventListener("scroll", onScroll, { passive: true });
        return () => window.removeEventListener("scroll", onScroll);
    }, []);

    return (
        <nav
            className={`fixed top-0 left-0 right-0 z-50 px-4 transition-all duration-300 ease-in-out ${
                scrolled ? "top-3" : "top-0"
            }`}
        >
            <div
                className={`mx-auto flex w-full items-center justify-between rounded-xl border px-4 transition-all duration-300 ease-in-out ${
                    scrolled
                        ? "max-w-4xl border-border bg-background/80 py-1.5 shadow-sm backdrop-blur-lg"
                        : "max-w-5xl border-transparent bg-transparent py-2 shadow-none backdrop-blur-none"
                }`}
            >
                <Link href="/" className="text-sm font-bold">
                    Online Judge
                </Link>
                <div className="flex items-center justify-center">
                    <Button variant="ghost" asChild>
                        <Link href="/explore">Explore</Link>
                    </Button>
                    <Button variant="ghost" asChild>
                        <Link href="/problems">Problems</Link>
                    </Button>
                    <Button variant="ghost" asChild>
                        <Link href="/pricing">Pricing</Link>
                    </Button>
                </div>
                <Show when="signed-out">
                    <div className="flex gap-2">
                        <SignInButton>
                            <Button>Sign In</Button>
                        </SignInButton>
                        <SignUpButton>
                            <Button variant="outline">Sign Up</Button>
                        </SignUpButton>
                    </div>
                </Show>
                <Show when="signed-in">
                    <div className="flex gap-2">
                        <div className="flex items-center justify-center">
                            <UserButton />
                        </div>
                        <Button variant="tertiary" asChild>
                            <Link href="/pricing">Premium</Link>
                        </Button>
                    </div>
                </Show>
            </div>
        </nav>
    );
};

export default Navbar;

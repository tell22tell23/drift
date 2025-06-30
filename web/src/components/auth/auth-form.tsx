"use client";

import { cn } from "@/lib/utils";
import { useState } from "react";

import { Link } from "@tanstack/react-router";

import { Github } from "@/components/icons/github";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { SignInForm } from "./sign-in-form";
import { SignUpForm } from "./sign-up-form";

interface SignInForm {
    redirectUrl?: string;
}

export function AuthForm({ redirectUrl = "/dashboard" } : SignInForm) {
    const [loading, setLoading] = useState(false);
    const [type, setType] = useState<"signin" | "signup">("signin");

    return (
        <Card className="max-w-80 border-none shadow-none bg-transparent text-center">
            <CardHeader>
                <CardTitle
                    className="text-2xl font-medium text-love"
                >
                    drift
                </CardTitle>
                <CardDescription
                    className="text-base"
                >
                    Write. Share. Vanish
                </CardDescription>
            </CardHeader>
            <CardContent>
                { type === "signin" ? (
                    <SignInForm redirectUrl={redirectUrl}/>
                ) : (
                        <SignUpForm setType={setType}/>
                    )
                }

                <Separator className="my-4" />

                <div className="grid gap-8">
                    <div
                        className={cn(
                            "flex w-full items-center gap-2",
                            "flex-col justify-between",
                        )}
                    >
                        <Button
                            variant="outline"
                            className={cn("w-full gap-2 cursor-pointer text-foreground")}
                            disabled={loading}
                            onClick={() => {
                                console.log("Loggin in")
                            }}
                        >
                            <Github className="fill-current" />
                            Continue with Github
                        </Button>
                    </div>
                </div>
            </CardContent>
            <CardFooter className="flex flex-col">
                <div className="text-center">
                    <p className="text-sm text-muted-foreground">
                        {type === "signin" ? "Don't have an account?" : "Already have an account?"}
                        {" "}
                        <p
                            className="underline text-foreground hover:text-rose cursor-pointer inline"
                            onClick={() => {
                                setType(type === "signin" ? "signup" : "signin");
                                setLoading(false);
                            }}
                        >
                            {type === "signin" ? "Sign Up" : "Sign In"}
                        </p>
                    </p>
                </div>
                <div className="flex w-full justify-center py-4">
                    <p className="text-center text-sm text-balance text-muted-foreground">
                        By continuing, you agree to our{" "}
                        <Link
                            to="/"
                            className="underline text-foreground hover:text-rose"
                        >
                            Terms of Service
                        </Link>{" "}
                        and{" "}
                        <Link
                            to="/"
                            className="underline text-foreground hover:text-rose"
                        >
                            Privacy Policy
                        </Link>
                    </p>
                </div>
            </CardFooter>
        </Card>
    );
}

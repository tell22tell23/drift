import React, { type SetStateAction } from "react";

import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form";

import { RiAlertLine } from "react-icons/ri";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useMutation } from "@tanstack/react-query";
import { SignUpFormSchema, type SignUp} from "@/types/auth";
import { auth } from "@/lib/auth/auth";

interface SignUpFormProps {
    setType: React.Dispatch<SetStateAction<"signin" | "signup">>;
};

export const SignUpForm = ({ setType }: SignUpFormProps) => {
    const [error, setError] = React.useState<string | null>(null);
    const form = useForm<SignUp>({
        resolver: zodResolver(SignUpFormSchema),
        defaultValues: {
            name: "",
            email: "",
            password: "",
        },
    });

    const { mutate, isPending } = useMutation({
        mutationFn: auth.signUp,
        onSuccess: () => setType("signin"),
        onError: () => setError("Failed to register. Please try again."),
    });

    const onSubmit = async (values: SignUp) => {
        setError(null);
        mutate(values);
    };

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                        <FormItem className="text-start">
                            <FormLabel
                                className="text-foreground"
                            >
                                Name
                            </FormLabel>
                            <FormControl>
                                <Input placeholder="John Doe" disabled={isPending} {...field} />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                        <FormItem className="text-start">
                            <FormLabel
                                className="text-foreground"
                            >
                                Email
                            </FormLabel>
                            <FormControl>
                                <Input placeholder="john.doe@0.com" disabled={isPending} {...field} />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                        <FormItem className="text-start">
                            <FormLabel
                                className="text-foreground"
                            >
                                Password
                            </FormLabel>
                            <FormControl>
                                <Input type="password" placeholder="******" disabled={isPending} {...field} />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                { error && (
                    <div className="text-destructive text-sm text-start flex items-start gap-2">
                        <RiAlertLine className="size-3 mt-1" />
                        <span>{error}</span>
                    </div>
                )}

                <Button
                    variant="secondary"
                    type="submit"
                    disabled={isPending}
                    className="w-full"
                >
                    {isPending ? "Signing Up..." : "Sign Up"}
                </Button>
            </form>
        </Form>
    );
}

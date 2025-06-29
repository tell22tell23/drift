import { useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form";
import { SignInFormSchema, type SignInType } from "@/types";
import { useNavigate } from '@tanstack/react-router';

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
import { RiAlertLine } from "react-icons/ri";
import { auth } from "@/lib/auth/auth";

interface SignInFormProps {
    redirectUrl: string;
};

export const SignInForm = ({
    redirectUrl,
}: SignInFormProps) => {
    const navigate = useNavigate();
    const [error, setError] = useState<string | null>(null);
    const form = useForm<SignInType>({
        resolver: zodResolver(SignInFormSchema),
        defaultValues: {
            email: "",
            password: "",
        },
    });

    const { mutate, isPending } = useMutation({
        mutationFn: auth.signIn,
        onSuccess: () => navigate({ to: redirectUrl }),
        onError: () => setError("Failed to sign in. Please check your credentials and try again."),
    });

    const onSubmit = async (values: SignInType) => {
        setError(null);
        mutate(values);
    };

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
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
                >Sign In</Button>
            </form>
        </Form>
    );
}

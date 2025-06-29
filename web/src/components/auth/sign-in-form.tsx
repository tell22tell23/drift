import React, { type SetStateAction } from "react";

import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form";
import { z } from "zod";

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

const SignInFormSchema = z.object({
    email: z.string().email("Invalid email address"),
    password: z.string().min(1, "Password is required"),
});
type SignInType = z.infer<typeof SignInFormSchema>;

interface SignInFormProps {
    redirectUrl: string;
    loading: boolean;
    setLoading: React.Dispatch<SetStateAction<boolean>>;
};

export const SignInForm = ({
    redirectUrl,
    loading,
    setLoading,
}: SignInFormProps) => {
    const form = useForm<SignInType>({
        resolver: zodResolver(SignInFormSchema),
        defaultValues: {
            email: "",
            password: "",
        },
    });

    const onSubmit = async (values: SignInType) => {
        //do some
        setLoading(true);
        console.log(values);
        console.log("Redirecting to:", redirectUrl);
        setLoading(false);
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
                                <Input placeholder="john.doe@0.com" disabled={loading} {...field} />
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
                                <Input type="password" placeholder="******" disabled={loading} {...field} />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <Button
                    variant="secondary"
                    type="submit"
                    disabled={loading}
                    className="w-full"
                >Sign In</Button>
            </form>
        </Form>
    );
}

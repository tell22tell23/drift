import { z } from "zod";

export const SignUpFormSchema = z.object({
    name: z.string().min(1, "Name is required"),
    email: z.string().email("Invalid email address"),
    password: z.string().min(6, "Password must be at least 6 character long"),
});
export type SignUpType = z.infer<typeof SignUpFormSchema>;

export const SignInFormSchema = z.object({
    email: z.string().email("Invalid email address"),
    password: z.string().min(1, "Password is required"),
});
export type SignInType = z.infer<typeof SignInFormSchema>;


import { z } from "zod";

export const SessionUserSchema = z.object({
    id: z.string(),
    email: z.string().email(),
    name: z.string(),
    image: z.string().nullable(),
    created_at: z.string(),
    updated_at: z.string(),
    expiresAt: z.string(),
});

export type SessionUser = z.infer<typeof SessionUserSchema>;

import dayjs from "dayjs"
import api from "@/lib/api";
import { queryClient } from "@/lib/react-query-client";
import { SESSION_QUERY_KEY } from "@/const";
import type { SignIn, SignUp } from "@/types/auth";
import { SessionUserSchema, type SessionUser } from "@/types/session";

export const auth = {
    async getSession(): Promise<SessionUser | null> {
        const cachedUser = queryClient.getQueryData<SessionUser>(SESSION_QUERY_KEY);
        const now = dayjs();

        if (cachedUser) {
            const expires = dayjs(cachedUser.expiresAt);
            if (expires.isAfter(now)) {
                return cachedUser;
            } else {
                queryClient.setQueryData(SESSION_QUERY_KEY, null);
                queryClient.invalidateQueries({ queryKey: SESSION_QUERY_KEY });
            }
        }

        try {
            const user = await queryClient.fetchQuery<SessionUser>({
                queryKey: ['session'],
                queryFn: async () => {
                    const res = await api.get("/users/me");
                    const parsed = SessionUserSchema.safeParse({
                        ...res.data.user,
                        expiresAt: res.data.expires_at,
                    });
                    if (!parsed.success) {
                        throw new Error("Invalid session data");
                    }
                    return parsed.data;
                },
            });
            return user;
        } catch (error: any) {
            if (error.response?.status === 401) {
                console.error("Her");
                queryClient.setQueryData(SESSION_QUERY_KEY, null);
                queryClient.invalidateQueries({ queryKey: SESSION_QUERY_KEY });
                return null;
            }
            throw error;
        }
    },

    async signUp(data: SignUp) {
        const res = await api.post("/auth/register", data);
        return res.data;
    },

    async signIn(data: SignIn) {
        const res = await api.post("/auth/login", data);
        return res.data;
    },

    async signOut() {
        const res = await api.post("/auth/logout");
        queryClient.setQueryData(SESSION_QUERY_KEY, null);
        queryClient.invalidateQueries({ queryKey: SESSION_QUERY_KEY });
        return res.data;
    }
}

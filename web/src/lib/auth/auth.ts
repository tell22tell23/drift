import api from "@/lib/api";
import type { SignInType, SignUpType } from "@/types";
import { queryClient } from "@/lib/react-query-client";

export const auth = {
    async getSession() {
        const cachedUser = queryClient.getQueryData(["session"]);
        if (cachedUser) {
            return cachedUser;
        }

        try {
            const user = await queryClient.fetchQuery({
                queryKey: ['session'],
                queryFn: async () => {
                    const res = await api.get("/users/me");
                    return {
                        ...res.data.user,
                        expiresAt: new Date(res.data.expires_at),
                    }
                },
            });
            return user;
        } catch (error: any) {
            if (error.response?.status === 401) {
                queryClient.setQueryData(["session"], null);
                return null;
            }
            throw error;
        }
    },

    async signUp(data: SignUpType) {
        const res = await api.post("/auth/register", data);
        return res.data;
    },

    async signIn(data: SignInType) {
        const res = await api.post("/auth/login", data);
        return res.data;
    },

    async signOut() {
        const res = await api.post("/auth/logout");
        queryClient.setQueryData(["session"], null);
        queryClient.invalidateQueries({ queryKey: ["session"] });
        return res.data;
    }
}

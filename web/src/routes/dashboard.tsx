import { auth } from "@/lib/auth/auth";
import { useMutation } from "@tanstack/react-query";
import {
    createFileRoute,
    redirect,
    useNavigate,
    useRouteContext,
} from "@tanstack/react-router";

import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute('/dashboard')({
    beforeLoad: async () => {
        const user = await auth.getSession();
        if (!user) {
            throw redirect({
                to: "/auth",
            });
        }
        return { user };
    },
    component: Dashboard
});

function Dashboard() {
    const navigate = useNavigate();
    const { user } = useRouteContext({ from: "/dashboard" });
    console.log("User:", user);

    const { mutate, isPending } = useMutation({
        mutationFn: auth.signOut,
        onSuccess: () => navigate({ to: "/auth" }),
        onError: (error) => {
            console.error("Error signing out user:", error);
        },
    });

    return (
        <section className="max-w-md mx-auto">
            <Header />
            <Button
                variant="secondary"
                className="mt-4 cursor-pointer"
                onClick={() => mutate()}
                disabled={isPending}
            >
                {isPending ? "Signing out..." : "Sign Out"}
            </Button>
        </section>
    );
}


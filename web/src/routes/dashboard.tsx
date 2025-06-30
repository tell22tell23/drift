import { auth } from "@/lib/auth/auth";
import {
    createFileRoute,
    redirect,
    useRouteContext,
} from "@tanstack/react-router";

import { Header } from "@/components/layout/header";
import { Separator } from "@/components/ui/separator";
import {
    DashboardHero,
    RepoBrowser,
} from "@/components/layout/dashboard";

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
    const { user } = useRouteContext({ from: "/dashboard" });

    return (
        <main className="max-w-2xl mx-auto">
            <Header className="mb-10" />

            <DashboardHero user={user} />

            <Separator className="mt-6 mb-4" />

            <RepoBrowser />
        </main>
    );
}


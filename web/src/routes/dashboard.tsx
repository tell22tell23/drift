import { Header } from "@/components/layout/header";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute('/dashboard')({
    component: Dashboard
});

function Dashboard() {
    return (
        <section className="max-w-md mx-auto">
            <Header />
        </section>
    );
}


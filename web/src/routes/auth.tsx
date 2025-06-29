import { AuthForm } from '@/components/auth/auth-form';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/auth')({
    component: Auth
});

function Auth() {
    return (
        <section className="flex h-dvh w-dvw items-center justify-center">
            <AuthForm />
        </section>
    );
}

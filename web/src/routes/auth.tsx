import { AuthForm } from '@/components/auth/auth-form';
import { auth } from '@/lib/auth/auth';
import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/auth')({
    beforeLoad: async () => {
        const user = await auth.getSession();
        if (user) {
            throw redirect({
                to: '/dashboard',
            });
        }
    },
    component: Auth
});

function Auth() {
    return (
        <section className="flex h-dvh w-dvw items-center justify-center">
            <AuthForm />
        </section>
    );
}

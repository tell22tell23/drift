import { IoMdMenu } from "react-icons/io";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { auth } from "@/lib/auth/auth";
import { useMutation } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { cn } from "@/lib/utils";

interface HeaderProps {
    className?: string;
}

export function Header({ className } : HeaderProps) {
    const navigate = useNavigate();

    const { mutate, isPending } = useMutation({
        mutationFn: auth.signOut,
        onSuccess: () => navigate({ to: "/auth" }),
        onError: (error) => {
            console.error("Sign out failed:", error);
        },
    });

    return (
        <nav className={cn("flex items-center justify-between py-4", className)}>
            <Link
                to="/dashboard"
                className="text-2xl text-love hover:text-love/80 transition-colors"
            >
                drift
            </Link>
            <DropdownMenu>
                <DropdownMenuTrigger className="cursor-pointer focus-visible:outline-none">
                    <IoMdMenu className="size-5 fill-love"/>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="end"
                    className="w-48"
                >
                    <DropdownMenuItem
                        className="cursor-pointer"
                        onClick={() => mutate()}
                        disabled={isPending}
                    >
                        Sign Out
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
        </nav>
    );
}

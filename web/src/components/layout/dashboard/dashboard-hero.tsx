import type { SessionUser } from "@/types/session";

import { Button } from "@/components/ui/button";
import { ArrowTopRight } from "@/components/icons/arrow-top-right";

interface DashboardHeroProps {
    user: SessionUser;
}

const getFirstName = (name: string): string => {
    const parts = name.toLowerCase().split(" ");
    return parts.length > 0 ? parts[0] : name;
}

export function DashboardHero({ user } : DashboardHeroProps) {
    return (
        <section>
            <div className="mb-8">
                <h1 className="mb-2">
                    hello,{" "}
                    <span className="text-foreground">{user.name}</span>
                    {" "}
                    <span className="text-muted-foreground text-base">&#123;draft@{getFirstName(user.name)}&#125;</span>
                </h1>
                {/* This is just placeholders now, need to work on it later */}
                <div className="flex items-center gap-x-2 text-xl">
                    <span className="text-green-700"> &gt; </span>
                    <span className="text-foreground">
                        Connected to 3 drifters
                    </span>
                </div>
                <div className="flex items-center gap-x-2 text-xl">
                    <span className="text-rose"> &gt; </span>
                    <span className="text-foreground">
                        Last synced 5 min ago
                    </span>
                </div>
            </div>
            <div className="flex items-center gap-x-4">
                <Button
                    variant="gold"
                    className="flex items-center justify-between text-md w-[150px]"
                >
                    <span>
                        New Draft
                    </span>
                    <ArrowTopRight className="size-3 fill-black" />
                </Button>

                <Button
                    variant="foam"
                    className="flex items-center justify-between text-md w-[150px]"
                >
                    <span>
                        Sync
                    </span>
                    <ArrowTopRight className="size-3 fill-black" />
                </Button>
            </div>

        </section>
    );
}

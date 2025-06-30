import { ArrowTopRight } from "@/components/icons/arrow-top-right";
import { Button } from "@/components/ui/button";
import { formatTimeAgo } from "@/hooks/use-date-time-formatter";

export interface Repo {
    id: string;
    title: string;
    description: string;
    owner: string;
    last_updated_by: string;
    created_at: string;
    updated_at: string;
}

interface RepoCardProps {
    repo: Repo
}

export function RepoCard({ repo }: RepoCardProps) {
    return (
        <div className="bg-midground border border-border p-4">
            <div className="mb-4">
                <h2 className="leading-tight">{repo.title}</h2>
                <p className="text-gold text-sm">dft:{repo.id}</p>
                <p className="text-md text-muted-foreground">
                    Last Updated:{" "}
                    <span className="text-foreground">
                        {formatTimeAgo(repo.updated_at)}
                    </span>
                    {" "}by{" "}
                    <span className="text-foreground">
                        {repo.last_updated_by}
                    </span>
                </p>
                <p className="text-iris text-base mt-2 leading-tight">{repo.description}</p>
            </div>

            <div className="flex items-center gap-x-2">
                <Button
                    variant="love"
                    className="flex items-center justify-between text-md w-[150px]"
                >
                    <span>
                        View
                    </span>
                    <ArrowTopRight className="size-3 fill-white" />
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
        </div>
    );
}

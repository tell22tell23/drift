"use client";

import { useState } from "react";

import { RepoCard, type Repo } from "./repo";
import { Input } from "@/components/ui/input";

const mockRepo: Repo[] = [{
    id: "123456",
    title: "Sample Repository",
    description: "This is a sample repository for testing purposes.",
    owner: "john_doe",
    last_updated_by: "jane_doe",
    created_at: "2023-10-01T12:00:00Z",
    updated_at: "2023-10-05T15:30:00Z"
}, {
    id: "789012",
    title: "Another Repository",
    description: "This is another repository for testing purposes.",
    owner: "alice_smith",
    last_updated_by: "bob_jones",
    created_at: "2023-09-15T08:00:00Z",
    updated_at: "2023-10-04T10:20:00Z"
}];

export function RepoBrowser() {
    const [searchQuery, setSearchQuery] = useState<string>("");

    const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (searchQuery.trim() === "") return;

        setSearchQuery("");

        // Here you would typically trigger a search operation,
    }

    return (
        <section>
            <form
                onSubmit={handleSearch}
                className="flex items-center bg-midground border border-border pl-2 text-base mb-8"
            >
                <div className="whitespace-nowrap flex items-center gap-x-2 text-iris">
                    <span>&gt;</span>
                    <span>search:</span>
                </div>
                <Input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="bg-transparent border-0 focus-visible:ring-0 focus-visible:border-0 focus-visible:outline-none text-md md:text-base"
                    autoComplete="off"
                />
            </form>
            <div className="grid grid-cols-2 gap-2">
                {mockRepo.map((repo) => (
                    <RepoCard key={repo.id} repo={repo} />
                ))}
            </div>
        </section>
    );
}

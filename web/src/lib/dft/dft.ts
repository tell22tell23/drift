import api from "@/lib/api";
import { z } from "zod";
import { queryClient } from "../react-query-client";

const DFTSchema = z.object({
    total: z.number(),
});

type DFTStats = z.infer<typeof DFTSchema>;

export const dft = {
    stats: async (dftID: string) => {
        const dft = await queryClient.fetchQuery<DFTStats>({
            queryKey: ['dft', 'stats', dftID],
            queryFn: async () => {
                const res = await api.get("/signal/stats?dft_id=" + dftID);
                const parsed = DFTSchema.safeParse({
                    total: res.data.total,
                });
                if (!parsed.success) {
                    throw new Error("Invalid session data");
                }
                return parsed.data;
            },
        });
        return dft ;
    },
}

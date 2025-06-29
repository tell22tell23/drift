import type { SignUpType } from "@/types";
import axios from "axios";

const api = axios.create({
    baseURL: import.meta.env.VITE_SERVER_BASE_URL!,
    withCredentials: true,
});

export const registerUser = async (data: SignUpType) => {
    const res = await api.post("/auth/register", data);
    return res.data;
};

export default api;

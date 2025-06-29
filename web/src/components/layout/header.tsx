import { IoMdMenu } from "react-icons/io";

export function Header() {
    return (
        <nav className="flex items-center justify-between py-4">
            <h1>drift</h1>
            <button
                className="cursor-pointer"
            >
                <IoMdMenu className="size-5 fill-love"/>
            </button>
        </nav>
    );
}

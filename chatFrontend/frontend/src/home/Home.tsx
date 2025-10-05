import { useState, useRef, useEffect } from "react";
import { LeftView } from "./Left";
import { RightView } from "./Right";
import { CheckIsUserOnline } from "../../wailsjs/go/main/App.js";
import "./Home.css";

type HomeProps = {
    user: string;
};

export default function SplitView({ user }: HomeProps) {
    const [leftW, setLeftW] = useState(300);
    const containerRef = useRef<HTMLDivElement | null>(null);
    const isResizing = useRef(false);

    const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
    const [selected, setSelected] = useState<string>(""); // current chat
    const [onlineMap, setOnlineMap] = useState<Record<string, boolean>>({});
    const isFetching = useRef(false);

    useEffect(() => {
        const interval = setInterval(async () => {
            if (isFetching.current || selectedUsers.length === 0) return;
            isFetching.current = true;

            try {
                const statusMap = await CheckIsUserOnline(selectedUsers);
                setOnlineMap(statusMap);
            } catch (err) {
                console.error(err);
            } finally {
                isFetching.current = false;
            }
        }, 1000);

        return () => clearInterval(interval);
    }, [selectedUsers]);

    const handleMouseDownDiv = () => { isResizing.current = true; };
    const handleMouseMoveDiv = (e: MouseEvent) => {
        if (!isResizing.current || !containerRef.current) return;
        const rect = containerRef.current.getBoundingClientRect();
        const newWidth = e.clientX - rect.left;
        if (newWidth > 300 && newWidth < rect.width - 300) setLeftW(newWidth);
    };
    const handleMouseUpDiv = () => { isResizing.current = false; };

    useEffect(() => {
        window.addEventListener("mousemove", handleMouseMoveDiv);
        window.addEventListener("mouseup", handleMouseUpDiv);
        return () => {
            window.removeEventListener("mousemove", handleMouseMoveDiv);
            window.removeEventListener("mouseup", handleMouseUpDiv);
        };
    }, []);

    return (
        <div ref={containerRef} className="split-container">
            <div style={{ width: `${leftW}px` }}>
                <LeftView
                    sender={user}
                    onSelect={(user) => {
                        setSelected(user);
                        if (!selectedUsers.includes(user)) {
                            setSelectedUsers(prev => [...prev, user]);
                        }
                    }}
                    onlineMap={onlineMap}
                />
            </div>

            <div className="divider" onMouseDown={handleMouseDownDiv} />

            <RightView
                selected={selected}
                sender={user}
                onlineMap={onlineMap}
            />
        </div>
    );
}

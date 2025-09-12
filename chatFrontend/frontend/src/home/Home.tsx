import React, { useState, useRef, useEffect } from "react";
import "./Home.css";
import { LeftView } from "./Left";
import { RightView } from "./Right";

type HomeProps = {
    user: string;
};

function SplitView({ user }: HomeProps) {
    const [leftW, setLeftW] = useState(300);
    const containerRef = useRef<HTMLDivElement | null>(null);
    const isResizing = useRef(false);
    const [selected, setselected] = useState("");

    const handleMouseDownDiv = () => {
        isResizing.current = true;
    };

    const handleMouseMoveDiv = (e: MouseEvent) => {
        if (!isResizing.current || !containerRef.current)
            return;

        const rect = containerRef.current.getBoundingClientRect();
        const newWidth = e.clientX - rect.left;

        if (newWidth > 300 && newWidth < rect.width - 300) {
            setLeftW(newWidth);
        }
    };

    const handleMouseUpDiv = () => {
        isResizing.current = false;
    };

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
                <LeftView onSelect={setselected}></LeftView>
            </div>

            <div className="divider" onMouseDown={handleMouseDownDiv} />

            <RightView selected={selected} sender={user}></RightView>
        </div>
    );
}

export default SplitView;

import { useState, useRef } from "react";
import { GetMessages, SearchUser as Search } from "../../wailsjs/go/main/App.js";
import { MessageHist } from "../types/MessageHist.js";

type LeftViewProps = {
    sender: string;
    onSelect: (username: string) => void;
    onlineMap: Record<string, boolean>;
};

export function LeftView({ sender, onSelect, onlineMap }: LeftViewProps) {
    const [username, setUsername] = useState("");
    const [results, setResults] = useState<string[]>([]);
    const [selected, setSelected] = useState<string[]>([]);
    const mpSelected = useRef<Record<string, boolean>>({});
    const [messages, setMessages] = useState<MessageHist[]>([]);

    const handleSearchUser = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!username.trim()) {
            setResults([]);
            return;
        }

        try {
            const res = await Search(username);
            if (res === "ok") {
                setResults([username]);
            }
            else {
                setResults(["User not found"]);
            }
        } catch (err: any) {
            setResults([`Error: ${err.toString()}`]);
        }
    };

    const handleSelectResult = async (value: string) => {
        if (mpSelected.current[value]) {
            setResults([]);
            return;
        }
        if (results[0] === "User not found") {
            setResults([]);
            onSelect("");
            return;
        }

        try {
            const msgs: MessageHist[] = await GetMessages(sender, value);
            setMessages(msgs);
        } catch (err) {
            console.error("Failed to fetch messages:", err);
        }

        setSelected(prev => [...prev, value]);
        mpSelected.current[value] = true;
        onSelect(value);
        setUsername("");
        setResults([]);
    };
    return (
        <div className="split-pane left-pane">
            <h2>Chats List</h2>

            <div className="search-wrapper">
                <form onSubmit={handleSearchUser} className="search-form">
                    <div className={`search ${results.length > 0 ? "has-results" : ""}`}>
                        <span className="material-symbols-outlined">search</span>
                        <input
                            className="search-input"
                            type="search"
                            placeholder="Search"
                            value={username}
                            onChange={(e) => {
                                const val = e.target.value;
                                setUsername(val);
                                if (!val.trim()) setResults([]);
                            }}
                        />
                        <button type="submit" style={{ display: "none" }} />
                    </div>

                    {results.length > 0 && (
                        <div className="search-dropdown">
                            {results.map((r, i) => (
                                <div key={i} className="search-item" onClick={() => handleSelectResult(r)}>
                                    {r}
                                </div>
                            ))}
                        </div>
                    )}
                </form>
            </div>

            <div className="chat-list">
                {selected.map((r, i) => (
                    <div key={i} className="chat-item" onClick={() => onSelect(r)}>
                        <div className="avatar">{r[0].toUpperCase()}</div>
                        <div className="chat-info">
                            <div className="chat-name">{r}</div>
                            <div className="chat-last">Last message preview...</div>
                            <div className="chat-temp">
                                <span
                                    className="chat-online"
                                    style={{ backgroundColor: onlineMap[r] ? "green" : "red" }}
                                />
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

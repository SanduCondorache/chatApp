import { useState } from "react";
import { SearchUser as Search } from "../../wailsjs/go/main/App.js";

type LeftViewProps = {
    onSelect: (value: string) => void;
};

export function LeftView({ onSelect }: LeftViewProps) {
    const [username, setUsername] = useState("");
    const [results, setResults] = useState<string[]>([]);
    const [selected, setselected] = useState<string[]>([]);

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
            } else {
                setResults(["User not found"]);
            }
        } catch (error: any) {
            setResults([`Error: ${error.toString()}`]);
        }
    };

    const handleSelectResult = (value: string) => {
        if (username === "clear") {
            setselected([]);
        }
        if (results[0] === "User not found") {
            setResults([]);
            onSelect("");
            return;
        }
        setselected(prev => [...prev, value]);
        onSelect(value);
        setUsername("");
        setResults([]);
    }


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
                                const value = e.target.value;
                                setUsername(value);

                                if (!value.trim()) {
                                    setResults([]);
                                }
                            }}
                        />
                        <button type="submit" style={{ display: "none" }} />
                    </div>
                    {results.length > 0 && (
                        <div className="search-dropdown">
                            {results.map((r, i) => (
                                <div
                                    key={i}
                                    className="search-item"
                                    onClick={() => handleSelectResult(r)}
                                >{r}</div>
                            ))}
                        </div>
                    )}
                </form>
            </div>
            {selected.map((r, i) => (
                <div key={i} className="selected-result">{r}</div>
            ))}
        </div>
    );
}

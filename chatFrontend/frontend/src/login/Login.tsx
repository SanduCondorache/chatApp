import { useNavigate } from "react-router-dom";
import "./Login.css";
import { Login as GoLogin } from "../../wailsjs/go/main/App.js"
import { useState } from "react";

type LoginProps = {
    onSelect: (value: string) => void;
};

export function Login({ onSelect }: LoginProps) {
    const navigate = useNavigate();
    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [err, setErr] = useState("");

    const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        try {
            const result = await GoLogin(username, email, password);
            if (result == "ok") {
                console.log("Login success");
                onSelect(username);
                navigate('/home');
            } else if (result == "username_taken") {
                setErr("Username is already taken");
            }
        } catch (error: any) {
            setErr("Error:" + error.toString());

        }

    };

    return (
        <section>
            <div className="login-box">
                <form onSubmit={handleLogin}>
                    <h2>Login</h2>
                    <div className="input-box">
                        <span className='icon'></span>
                        <input type="username" value={username} onChange={e => setUsername(e.target.value)} required />
                        <label>Username</label>
                    </div>
                    <div className="input-box">
                        <span className='icon'></span>
                        <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
                        <label>Email</label>
                    </div>
                    <div className="input-box">
                        <span className='icon'></span>
                        <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
                        <label>Password</label>
                    </div>
                    {err && <p className="error-message">{err}</p>}
                    <button type='submit'>Login</button>
                    <div className='register-link'>
                        <p> Don't have an account? <a href='#'>Register</a></p>
                    </div>
                </form>
            </div>
        </section>
    )

}

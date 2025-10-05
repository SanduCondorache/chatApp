import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Home from './home/Home';
import { Login } from './login/Login';
import { Register } from './register/Register'

function App() {
    const [user, setuser] = useState("");
    return (
        <Router>
            <Routes>
                <Route path="/" element={<Login onSelect={setuser} />} />
                <Route path="/home" element={<Home user={user} />} />
                <Route path="/register" element={<Register />} />
            </Routes>
        </Router>
    );
}

export default App;

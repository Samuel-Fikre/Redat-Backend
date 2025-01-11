import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Navbar from './components/Navbar';
import StationList from './pages/StationList';
import RouteList from './pages/RouteList';

function App() {
  return (
    <Router>
      <div className="app">
        <Navbar />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<StationList />} />
            <Route path="/stations" element={<StationList />} />
            <Route path="/routes" element={<RouteList />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App; 
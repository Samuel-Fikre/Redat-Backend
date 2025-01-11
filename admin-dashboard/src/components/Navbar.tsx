import { Link } from 'react-router-dom';

const Navbar = () => {
  return (
    <nav className="navbar">
      <h1>Taxi Admin</h1>
      <ul className="nav-links">
        <li>
          <Link to="/stations">Stations</Link>
        </li>
        <li>
          <Link to="/routes">Routes</Link>
        </li>
      </ul>
    </nav>
  );
};

export default Navbar; 
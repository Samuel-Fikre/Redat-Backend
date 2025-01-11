import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface Station {
  id: string;
  name: string;
  location: {
    type: string;
    coordinates: [number, number];
  };
}

interface ApiResponse {
  stations: Station[];
}

const StationList = () => {
  const [stations, setStations] = useState<Station[]>([]);
  const [error, setError] = useState<string>('');
  const [newStation, setNewStation] = useState({
    name: '',
    latitude: '',
    longitude: '',
  });
  const [editingStation, setEditingStation] = useState<Station | null>(null);

  useEffect(() => {
    fetchStations();
  }, []);

  const fetchStations = async () => {
    try {
      const response = await axios.get<ApiResponse>('http://localhost:8080/stations');
      if (response.data && Array.isArray(response.data.stations)) {
        setStations(response.data.stations);
        setError('');
      } else {
        setStations([]);
        setError('Invalid data format received from server');
      }
    } catch (error) {
      console.error('Error fetching stations:', error);
      setStations([]);
      setError('Failed to fetch stations');
    }
  };

  const handleAddStation = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await axios.post('http://localhost:8080/stations', {
        name: newStation.name,
        location: {
          type: 'Point',
          coordinates: [Number(newStation.longitude), Number(newStation.latitude)],
        },
      });
      if (response.status === 201) {
        setNewStation({ name: '', latitude: '', longitude: '' });
        setError('');
        fetchStations();
      }
    } catch (error) {
      console.error('Error adding station:', error);
      setError('Failed to add station');
    }
  };

  const handleUpdateStation = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingStation) return;

    try {
      const response = await axios.put(`http://localhost:8080/stations/${editingStation.id}`, {
        name: editingStation.name,
        location: editingStation.location,
      });
      if (response.status === 200) {
        setEditingStation(null);
        setError('');
        fetchStations();
      }
    } catch (error) {
      console.error('Error updating station:', error);
      setError('Failed to update station');
    }
  };

  const handleDeleteStation = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this station?')) {
      return;
    }

    try {
      const response = await axios.delete(`http://localhost:8080/stations/${id}`);
      if (response.status === 200) {
        setError('');
        fetchStations();
      }
    } catch (error: any) {
      console.error('Error deleting station:', error);
      if (error.response) {
        // Server responded with an error
        if (error.response.status === 404) {
          setError('Station not found');
        } else if (error.response.status === 500) {
          setError('Cannot delete station: it might be referenced by existing routes');
        } else {
          setError(error.response.data?.error || 'Failed to delete station');
        }
      } else if (error.request) {
        // Request was made but no response
        setError('No response from server. Please check your connection.');
      } else {
        // Something else went wrong
        setError('An error occurred while deleting the station');
      }
    }
  };

  return (
    <div>
      <h2>Stations</h2>
      {error && <div className="error-message" style={{ color: 'red', margin: '10px 0' }}>{error}</div>}
      
      <form onSubmit={handleAddStation} className="form-group">
        <h3>Add New Station</h3>
        <div>
          <label>Name:</label>
          <input
            type="text"
            className="form-control"
            value={newStation.name}
            onChange={(e) => setNewStation({ ...newStation, name: e.target.value })}
            required
          />
        </div>
        <div>
          <label>Latitude:</label>
          <input
            type="number"
            step="any"
            className="form-control"
            value={newStation.latitude}
            onChange={(e) => setNewStation({ ...newStation, latitude: e.target.value })}
            required
          />
        </div>
        <div>
          <label>Longitude:</label>
          <input
            type="number"
            step="any"
            className="form-control"
            value={newStation.longitude}
            onChange={(e) => setNewStation({ ...newStation, longitude: e.target.value })}
            required
          />
        </div>
        <button type="submit" className="button button-primary">Add Station</button>
      </form>

      {stations.length === 0 ? (
        <p>No stations found</p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Latitude</th>
              <th>Longitude</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {stations.map((station) => (
              <tr key={station.id}>
                <td>{station.name}</td>
                <td>{station.location.coordinates[1]}</td>
                <td>{station.location.coordinates[0]}</td>
                <td>
                  <button
                    className="button button-primary"
                    onClick={() => setEditingStation(station)}
                  >
                    Edit
                  </button>
                  <button
                    className="button button-danger"
                    onClick={() => handleDeleteStation(station.id)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {editingStation && (
        <div className="modal">
          <form onSubmit={handleUpdateStation} className="form-group">
            <h3>Edit Station</h3>
            <div>
              <label>Name:</label>
              <input
                type="text"
                className="form-control"
                value={editingStation.name}
                onChange={(e) =>
                  setEditingStation({
                    ...editingStation,
                    name: e.target.value,
                  })
                }
                required
              />
            </div>
            <div>
              <label>Latitude:</label>
              <input
                type="number"
                step="any"
                className="form-control"
                value={editingStation.location.coordinates[1]}
                onChange={(e) =>
                  setEditingStation({
                    ...editingStation,
                    location: {
                      ...editingStation.location,
                      coordinates: [
                        editingStation.location.coordinates[0],
                        Number(e.target.value),
                      ],
                    },
                  })
                }
                required
              />
            </div>
            <div>
              <label>Longitude:</label>
              <input
                type="number"
                step="any"
                className="form-control"
                value={editingStation.location.coordinates[0]}
                onChange={(e) =>
                  setEditingStation({
                    ...editingStation,
                    location: {
                      ...editingStation.location,
                      coordinates: [
                        Number(e.target.value),
                        editingStation.location.coordinates[1],
                      ],
                    },
                  })
                }
                required
              />
            </div>
            <button type="submit" className="button button-primary">Update Station</button>
            <button
              type="button"
              className="button button-danger"
              onClick={() => setEditingStation(null)}
            >
              Cancel
            </button>
          </form>
        </div>
      )}
    </div>
  );
};

export default StationList; 
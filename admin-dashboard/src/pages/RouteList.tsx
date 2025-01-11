import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface Station {
  id: string;
  name: string;
}

interface Route {
  id: string;
  from: string;
  to: string;
  price: number;
  isDirectRoute: boolean;
  intermediateStations?: string[]; // Names of stations in between
}

interface ApiResponse {
  routes: Route[];
}

const RouteList = () => {
  const [routes, setRoutes] = useState<Route[]>([]);
  const [stations, setStations] = useState<Station[]>([]);
  const [error, setError] = useState<string>('');
  const [newRoute, setNewRoute] = useState({
    startStationId: '',
    endStationId: '',
    price: '',
    isDirectRoute: true,
    intermediateStations: [] as string[],
  });
  const [editingRoute, setEditingRoute] = useState<Route | null>(null);
  const [showIntermediateStations, setShowIntermediateStations] = useState(false);

  useEffect(() => {
    fetchRoutes();
    fetchStations();
  }, []);

  const fetchRoutes = async () => {
    try {
      const response = await axios.get<ApiResponse>('http://localhost:8080/routes');
      if (response.data && Array.isArray(response.data.routes)) {
        setRoutes(response.data.routes);
        setError('');
      } else {
        setRoutes([]);
        setError('Invalid data format received from server');
      }
    } catch (error) {
      console.error('Error fetching routes:', error);
      setRoutes([]);
      setError('Failed to fetch routes');
    }
  };

  const fetchStations = async () => {
    try {
      const response = await axios.get('http://localhost:8080/stations');
      if (response.data && Array.isArray(response.data.stations)) {
        setStations(response.data.stations);
      }
    } catch (error) {
      console.error('Error fetching stations:', error);
    }
  };

  const handleAddRoute = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const routeData = {
        from: newRoute.startStationId,
        to: newRoute.endStationId,
        price: Number(newRoute.price),
        isDirectRoute: newRoute.isDirectRoute,
        intermediateStations: newRoute.isDirectRoute ? [] : newRoute.intermediateStations,
      };

      const response = await axios.post('http://localhost:8080/routes', routeData);
      if (response.status === 201) {
        setNewRoute({
          startStationId: '',
          endStationId: '',
          price: '',
          isDirectRoute: true,
          intermediateStations: [],
        });
        setShowIntermediateStations(false);
        setError('');
        fetchRoutes();
      }
    } catch (error: any) {
      console.error('Error adding route:', error);
      setError(error.response?.data?.error || 'Failed to add route');
    }
  };

  const handleUpdateRoute = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingRoute) return;

    try {
      const response = await axios.put(`http://localhost:8080/routes/${editingRoute.id}`, {
        from: editingRoute.from,
        to: editingRoute.to,
        price: editingRoute.price,
        isDirectRoute: editingRoute.isDirectRoute,
        intermediateStations: editingRoute.intermediateStations || [],
      });
      if (response.status === 200) {
        setEditingRoute(null);
        setError('');
        fetchRoutes();
      }
    } catch (error: any) {
      console.error('Error updating route:', error);
      setError(error.response?.data?.error || 'Failed to update route');
    }
  };

  const handleDeleteRoute = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this route?')) {
      return;
    }

    try {
      const response = await axios.delete(`http://localhost:8080/routes/${id}`);
      if (response.status === 200) {
        setError('');
        fetchRoutes();
      }
    } catch (error: any) {
      console.error('Error deleting route:', error);
      setError(error.response?.data?.error || 'Failed to delete route');
    }
  };

  const handleAddIntermediateStation = () => {
    if (!newRoute.isDirectRoute) {
      setNewRoute({
        ...newRoute,
        intermediateStations: [...newRoute.intermediateStations, ''],
      });
    }
  };

  const handleRemoveIntermediateStation = (index: number) => {
    const updatedStations = [...newRoute.intermediateStations];
    updatedStations.splice(index, 1);
    setNewRoute({
      ...newRoute,
      intermediateStations: updatedStations,
    });
  };

  const handleIntermediateStationChange = (index: number, value: string) => {
    const updatedStations = [...newRoute.intermediateStations];
    updatedStations[index] = value;
    setNewRoute({
      ...newRoute,
      intermediateStations: updatedStations,
    });
  };

  return (
    <div>
      <h2>Routes</h2>
      {error && <div className="error-message" style={{ color: 'red', margin: '10px 0' }}>{error}</div>}
      
      <form onSubmit={handleAddRoute} className="form-group">
        <h3>Add New Route</h3>
        <div>
          <label>Start Station:</label>
          <select
            className="form-control"
            value={newRoute.startStationId}
            onChange={(e) => setNewRoute({ ...newRoute, startStationId: e.target.value })}
            required
          >
            <option value="">Select Start Station</option>
            {stations.map((station) => (
              <option key={station.id} value={station.name}>
                {station.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label>End Station:</label>
          <select
            className="form-control"
            value={newRoute.endStationId}
            onChange={(e) => setNewRoute({ ...newRoute, endStationId: e.target.value })}
            required
          >
            <option value="">Select End Station</option>
            {stations.map((station) => (
              <option key={station.id} value={station.name}>
                {station.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label>
            <input
              type="checkbox"
              checked={!newRoute.isDirectRoute}
              onChange={(e) => {
                setNewRoute({
                  ...newRoute,
                  isDirectRoute: !e.target.checked,
                  intermediateStations: !e.target.checked ? [] : newRoute.intermediateStations,
                });
                setShowIntermediateStations(e.target.checked);
              }}
            />
            Has Intermediate Stations
          </label>
        </div>
        {showIntermediateStations && (
          <div>
            <label>Intermediate Stations:</label>
            {newRoute.intermediateStations.map((station, index) => (
              <div key={index} style={{ display: 'flex', gap: '10px', marginBottom: '10px' }}>
                <select
                  className="form-control"
                  value={station}
                  onChange={(e) => handleIntermediateStationChange(index, e.target.value)}
                  required
                >
                  <option value="">Select Station</option>
                  {stations.map((s) => (
                    <option key={s.id} value={s.name}>
                      {s.name}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  className="button button-danger"
                  onClick={() => handleRemoveIntermediateStation(index)}
                >
                  Remove
                </button>
              </div>
            ))}
            <button
              type="button"
              className="button button-secondary"
              onClick={handleAddIntermediateStation}
            >
              Add Intermediate Station
            </button>
          </div>
        )}
        <div>
          <label>Price:</label>
          <input
            type="number"
            step="0.01"
            className="form-control"
            value={newRoute.price}
            onChange={(e) => setNewRoute({ ...newRoute, price: e.target.value })}
            required
          />
        </div>
        <button type="submit" className="button button-primary">Add Route</button>
      </form>

      {routes.length === 0 ? (
        <p>No routes found</p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>From</th>
              <th>To</th>
              <th>Via</th>
              <th>Price</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {routes.map((route) => (
              <tr key={route.id}>
                <td>{route.from}</td>
                <td>{route.to}</td>
                <td>
                  {route.intermediateStations && route.intermediateStations.length > 0
                    ? route.intermediateStations.join(' → ')
                    : 'Direct Route'}
                </td>
                <td>{route.price}</td>
                <td>
                  <button
                    className="button button-primary"
                    onClick={() => setEditingRoute(route)}
                  >
                    Edit
                  </button>
                  <button
                    className="button button-danger"
                    onClick={() => handleDeleteRoute(route.id)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {editingRoute && (
        <div className="modal">
          <form onSubmit={handleUpdateRoute} className="form-group">
            <h3>Edit Route</h3>
            <div>
              <label>Start Station:</label>
              <select
                className="form-control"
                value={editingRoute.from}
                onChange={(e) =>
                  setEditingRoute({
                    ...editingRoute,
                    from: e.target.value,
                  })
                }
                required
              >
                {stations.map((station) => (
                  <option key={station.id} value={station.name}>
                    {station.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label>End Station:</label>
              <select
                className="form-control"
                value={editingRoute.to}
                onChange={(e) =>
                  setEditingRoute({
                    ...editingRoute,
                    to: e.target.value,
                  })
                }
                required
              >
                {stations.map((station) => (
                  <option key={station.id} value={station.name}>
                    {station.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label>
                <input
                  type="checkbox"
                  checked={!editingRoute.isDirectRoute}
                  onChange={(e) =>
                    setEditingRoute({
                      ...editingRoute,
                      isDirectRoute: !e.target.checked,
                      intermediateStations: !e.target.checked ? [] : editingRoute.intermediateStations,
                    })
                  }
                />
                Has Intermediate Stations
              </label>
            </div>
            {!editingRoute.isDirectRoute && (
              <div>
                <label>Intermediate Stations:</label>
                {(editingRoute.intermediateStations || []).map((station, index) => (
                  <div key={index} style={{ display: 'flex', gap: '10px', marginBottom: '10px' }}>
                    <select
                      className="form-control"
                      value={station}
                      onChange={(e) => {
                        const updatedStations = [...(editingRoute.intermediateStations || [])];
                        updatedStations[index] = e.target.value;
                        setEditingRoute({
                          ...editingRoute,
                          intermediateStations: updatedStations,
                        });
                      }}
                      required
                    >
                      <option value="">Select Station</option>
                      {stations.map((s) => (
                        <option key={s.id} value={s.name}>
                          {s.name}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      className="button button-danger"
                      onClick={() => {
                        const updatedStations = [...(editingRoute.intermediateStations || [])];
                        updatedStations.splice(index, 1);
                        setEditingRoute({
                          ...editingRoute,
                          intermediateStations: updatedStations,
                        });
                      }}
                    >
                      Remove
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  className="button button-secondary"
                  onClick={() =>
                    setEditingRoute({
                      ...editingRoute,
                      intermediateStations: [...(editingRoute.intermediateStations || []), ''],
                    })
                  }
                >
                  Add Intermediate Station
                </button>
              </div>
            )}
            <div>
              <label>Price:</label>
              <input
                type="number"
                step="0.01"
                className="form-control"
                value={editingRoute.price}
                onChange={(e) =>
                  setEditingRoute({
                    ...editingRoute,
                    price: Number(e.target.value),
                  })
                }
                required
              />
            </div>
            <button type="submit" className="button button-primary">Update Route</button>
            <button
              type="button"
              className="button button-danger"
              onClick={() => setEditingRoute(null)}
            >
              Cancel
            </button>
          </form>
        </div>
      )}
    </div>
  );
};

export default RouteList; 
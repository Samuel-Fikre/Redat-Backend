import React from 'react';

interface RouteLeg {
  from: string;
  to: string;
  price: number;
}

interface JourneyDetailsProps {
  route: string[];
  totalPrice: number;
  legs: RouteLeg[];
}

const JourneyDetails: React.FC<JourneyDetailsProps> = ({ route, totalPrice, legs }) => {
  return (
    <div className="journey-details">
      <h2>Journey Details</h2>
      <div className="journey-info">
        <p><strong>Total Price:</strong> {totalPrice} Birr</p>
        
        <div className="route-details">
          <h3>Route Information:</h3>
          <div className="route-path">
            {route.map((station, index) => (
              <React.Fragment key={index}>
                <span>{station}</span>
                {index < route.length - 1 && <span className="arrow">→</span>}
              </React.Fragment>
            ))}
          </div>
          
          {legs.length > 1 && (
            <div className="route-legs">
              <h4>Journey Segments:</h4>
              {legs.map((leg, index) => (
                <div key={index} className="leg-info">
                  <p>
                    {leg.from} → {leg.to}: {leg.price} Birr
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default JourneyDetails;

// Add some CSS styles
const styles = `
.journey-details {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  margin: 20px 0;
}

.journey-info {
  margin-top: 15px;
}

.route-details {
  margin-top: 20px;
}

.route-path {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 15px 0;
  flex-wrap: wrap;
}

.arrow {
  color: #666;
}

.route-legs {
  margin-top: 15px;
  padding: 15px;
  background-color: #f5f5f5;
  border-radius: 4px;
}

.leg-info {
  margin: 10px 0;
}

.leg-info p {
  margin: 5px 0;
}
`;

// Create a style element and append it to the document head
const styleSheet = document.createElement("style");
styleSheet.innerText = styles;
document.head.appendChild(styleSheet); 
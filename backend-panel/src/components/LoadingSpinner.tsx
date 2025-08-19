import React from 'react';
import '../styles/LoadingSpinner.css';

interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  color?: 'primary' | 'white';
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ 
  size = 'medium', 
  color = 'primary' 
}) => {
  return (
    <div className={`loading-spinner ${size} ${color}`}>
      <div className="spinner-circle"></div>
    </div>
  );
};

export default LoadingSpinner;

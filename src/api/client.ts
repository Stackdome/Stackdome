// Generic axios API client setup
import axios from 'axios';

const api = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '/api/v1'),
  headers: {
    'Content-Type': 'application/json',
  },
});

// Optional: Add a response interceptor for better error handling
api.interceptors.response.use(
  response => response,
  error => {
    // You can customize this logic as needed
    if (error.response) {
      // Server responded with a status other than 2xx
      return Promise.reject(error.response.data || error);
    } else if (error.request) {
      // No response received
      return Promise.reject({ message: 'Network error. Please try again.' });
    } else {
      // Something else happened
      return Promise.reject({ message: error.message });
    }
  }
);

export default api;

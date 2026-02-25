// Simple authentication helper
// In a real app, this would connect to your backend API

export const authHelpers = {
  // Login function - sets a token in localStorage
  login: (token: string) => {
    localStorage.setItem("authToken", token);
  },

  // Logout function - removes token
  logout: () => {
    localStorage.removeItem("authToken");
  },

  // Check if user is authenticated
  isAuthenticated: () => {
    return !!localStorage.getItem("authToken");
  },

  // Get current token
  getToken: () => {
    return localStorage.getItem("authToken");
  },
};

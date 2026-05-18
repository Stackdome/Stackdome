import { useTheme as useThemeContext } from "@/contexts/theme-provider";

/**
 * A hook to access the theme context
 *
 * @returns {object} An object with theme and setTheme
 */
export const useTheme = useThemeContext;

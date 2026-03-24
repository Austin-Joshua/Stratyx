import { useAuthContext } from "@/context/authContext";

export function useAuth() {
  return useAuthContext();
}

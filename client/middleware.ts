import { authMiddleware } from "@clerk/nextjs";


export default authMiddleware({
 publicRoutes: ["/"],
 ignoredRoutes: ["/about"]
});

export const config = {
  matcher: ["/((?!.*\\..*|_next).*)", "/", "/(api|trpc)(.*)"],
};

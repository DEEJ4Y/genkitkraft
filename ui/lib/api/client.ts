import createFetchClient from "openapi-fetch"
import createClient from "openapi-react-query"
import type { paths } from "./schema"

const fetchClient = createFetchClient<paths>({
  baseUrl: process.env.NEXT_PUBLIC_API_BASE || "",
  credentials: "include",
})

export { fetchClient }

export const $api = createClient(fetchClient)

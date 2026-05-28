import { useParams } from "react-router-dom"
import { TicketsList } from "../components/TicketsList"
import { TicketDetail } from "../components/TicketDetail"

export function TicketsListRoute() {
  return <TicketsList />
}

export function TicketDetailRoute() {
  const { id } = useParams()
  if (!id) return null
  return <TicketDetail ticketId={id} />
}

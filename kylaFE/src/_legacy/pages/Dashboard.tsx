import { ChartAreaInteractive } from "../components/chart-area-interactive";
import { DataTable } from "../components/data-table";
import { SectionCards } from "../components/section-cards";
import { data } from "../app/dashboard/data";
import PageContainer from "@/components/page-container";

export default function Page() {
  return (
    <PageContainer>
      <SectionCards />
      <ChartAreaInteractive />
      <DataTable data={data} />
    </PageContainer>
  );
}

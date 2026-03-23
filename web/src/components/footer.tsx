import { Wheat } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t bg-muted/30 mt-auto">
      <div className="mx-auto max-w-7xl px-4 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Wheat className="h-5 w-5 text-primary" />
            <span className="font-semibold text-primary">SanGiaGao.Vn</span>
          </div>
          <p className="text-sm text-muted-foreground">
            Sàn giao gạo trực tuyến &copy; {new Date().getFullYear()}
          </p>
        </div>
      </div>
    </footer>
  );
}

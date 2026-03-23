"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, MapPin, Package, MessageCircle, Star, Flag } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { getListingDetail, createConversation, createReport, getMessages, sendMessage, type ListingDetail } from "@/services/api";
import { formatPrice, formatQuantity, formatDate, timeAgo } from "@/lib/utils";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

export default function ListingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user, token } = useAuth();
  const router = useRouter();
  const [listing, setListing] = useState<ListingDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedImage, setSelectedImage] = useState(0);
  const [contacting, setContacting] = useState(false);

  // Report state
  const [showReport, setShowReport] = useState(false);
  const [reportReason, setReportReason] = useState("");
  const [reportDesc, setReportDesc] = useState("");
  const [submittingReport, setSubmittingReport] = useState(false);

  useEffect(() => {
    if (id) {
      getListingDetail(id)
        .then(setListing)
        .catch(() => toast.error("Không tìm thấy tin đăng"))
        .finally(() => setLoading(false));
    }
  }, [id]);

  async function handleContact() {
    if (!token || !user) {
      router.push("/dang-nhap");
      return;
    }
    if (!listing?.seller || !id) return;
    setContacting(true);
    try {
      const conv = await createConversation(token, listing.seller.id, listing.id);
      // Auto-send listing_link if not already sent today (like mobile)
      try {
        const msgs = await getMessages(token, conv.id, 1, 30);
        const today = new Date().toDateString();
        const alreadySent = msgs.data.some(
          (m) => m.type === "listing_link" && m.content === `listing://${id}` && new Date(m.created_at).toDateString() === today
        );
        if (!alreadySent) {
          await sendMessage(token, conv.id, `listing://${id}`, "listing_link");
        }
      } catch {
        // ignore - still navigate to chat
      }
      router.push(`/tin-nhan/${conv.id}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Không thể liên hệ");
    } finally {
      setContacting(false);
    }
  }

  async function handleReport(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !id) {
      router.push("/dang-nhap");
      return;
    }
    if (!reportReason) {
      toast.error("Vui lòng chọn lý do");
      return;
    }
    setSubmittingReport(true);
    try {
      await createReport(token, "listing", id, reportReason, reportDesc || undefined);
      toast.success("Đã gửi báo cáo");
      setShowReport(false);
      setReportReason("");
      setReportDesc("");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Báo cáo thất bại");
    } finally {
      setSubmittingReport(false);
    }
  }

  if (loading) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-6 space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-80 w-full rounded-lg" />
        <Skeleton className="h-6 w-3/4" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  if (!listing) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-12 text-center">
        <p className="text-muted-foreground">Tin đăng không tồn tại hoặc đã bị xóa</p>
        <Link href="/san-giao-dich">
          <Button variant="outline" className="mt-4">Quay lại sàn</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <Link href="/san-giao-dich" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="h-4 w-4" />
        Quay lại
      </Link>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Images + Details */}
        <div className="lg:col-span-2 space-y-4">
          {/* Image Gallery */}
          <div className="rounded-lg overflow-hidden bg-muted aspect-video flex items-center justify-center">
            {listing.images.length > 0 ? (
              <img
                src={listing.images[selectedImage]}
                alt={listing.title}
                className="max-h-full max-w-full object-contain"
              />
            ) : (
              <Package className="h-16 w-16 text-muted-foreground/40" />
            )}
          </div>
          {listing.images.length > 1 && (
            <div className="flex gap-2 overflow-x-auto">
              {listing.images.map((img, i) => (
                <button
                  key={i}
                  onClick={() => setSelectedImage(i)}
                  className={`h-16 w-16 rounded-md overflow-hidden border-2 flex-shrink-0 ${
                    i === selectedImage ? "border-primary" : "border-transparent"
                  }`}
                >
                  <img src={img} alt="" className="h-full w-full object-cover" />
                </button>
              ))}
            </div>
          )}

          {/* Details */}
          <Card>
            <CardContent className="p-6">
              <h1 className="text-xl font-bold mb-2">{listing.title}</h1>
              <p className="text-2xl font-bold text-primary mb-4">
                {formatPrice(listing.price_per_kg)}
              </p>

              <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                <div>
                  <span className="text-muted-foreground">Loại gạo:</span>
                  <p className="font-medium">{listing.rice_type}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Số lượng:</span>
                  <p className="font-medium">{formatQuantity(listing.quantity_kg)}</p>
                </div>
                {listing.province && (
                  <div>
                    <span className="text-muted-foreground">Khu vực:</span>
                    <p className="font-medium flex items-center gap-1">
                      <MapPin className="h-3.5 w-3.5" />
                      {listing.province}
                      {listing.district && `, ${listing.district}`}
                    </p>
                  </div>
                )}
                {listing.harvest_season && (
                  <div>
                    <span className="text-muted-foreground">Vụ mùa:</span>
                    <p className="font-medium">{listing.harvest_season}</p>
                  </div>
                )}
              </div>

              {listing.description && (
                <div className="border-t pt-4">
                  <h3 className="font-medium mb-2">Mô tả</h3>
                  <p className="text-sm text-muted-foreground whitespace-pre-wrap">
                    {listing.description}
                  </p>
                </div>
              )}

              {listing.certifications && (
                <div className="border-t pt-4 mt-4">
                  <h3 className="font-medium mb-2">Chứng nhận</h3>
                  <p className="text-sm text-muted-foreground">{listing.certifications}</p>
                </div>
              )}

              <div className="border-t pt-4 mt-4 flex items-center gap-4 text-xs text-muted-foreground">
                <span>Đăng: {formatDate(listing.created_at)}</span>
                <span>{listing.view_count} lượt xem</span>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Seller Sidebar */}
        <div className="space-y-4">
          {listing.seller && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Người bán
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Link href={`/nguoi-ban/${listing.seller.id}`} className="block mb-4 hover:opacity-80">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-12 w-12">
                      <AvatarFallback className="bg-primary/10 text-primary font-semibold">
                        {(listing.seller.name || "?").charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-semibold">{listing.seller.name || "Ẩn danh"}</p>
                      {listing.seller.org_name && (
                        <p className="text-xs text-muted-foreground">{listing.seller.org_name}</p>
                      )}
                      {listing.seller.province && (
                        <p className="text-xs text-muted-foreground flex items-center gap-1">
                          <MapPin className="h-3 w-3" />
                          {listing.seller.province}
                        </p>
                      )}
                    </div>
                  </div>
                </Link>
                <p className="text-xs text-muted-foreground mb-4">
                  Thành viên từ {formatDate(listing.seller.created_at)}
                </p>
                <Button
                  className="w-full gap-2"
                  onClick={handleContact}
                  disabled={contacting || listing.seller.id === user?.id}
                >
                  <MessageCircle className="h-4 w-4" />
                  {contacting ? "Đang xử lý..." : "Chat với người bán"}
                </Button>
              </CardContent>
            </Card>
          )}

          <Card>
            <CardContent className="p-4 flex flex-col gap-2">
              {listing.seller && (
                <Link href={`/nguoi-ban/${listing.seller.id}`}>
                  <Button variant="outline" size="sm" className="gap-2 justify-start w-full">
                    <Star className="h-4 w-4" />
                    Xem đánh giá người bán
                  </Button>
                </Link>
              )}
              <Button
                variant="outline"
                size="sm"
                className="gap-2 justify-start text-destructive"
                onClick={() => setShowReport(!showReport)}
              >
                <Flag className="h-4 w-4" />
                Báo cáo tin đăng
              </Button>
            </CardContent>
          </Card>

          {/* Report form */}
          {showReport && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium">Báo cáo tin đăng</CardTitle>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleReport} className="space-y-3">
                  <select
                    value={reportReason}
                    onChange={(e) => setReportReason(e.target.value)}
                    className="w-full h-10 rounded-md border border-input bg-background px-3 text-sm"
                    required
                  >
                    <option value="">Chọn lý do</option>
                    <option value="fraud">Lừa đảo</option>
                    <option value="false_info">Thông tin sai lệch</option>
                    <option value="spam">Spam</option>
                    <option value="other">Khác</option>
                  </select>
                  <textarea
                    value={reportDesc}
                    onChange={(e) => setReportDesc(e.target.value)}
                    placeholder="Mô tả chi tiết (không bắt buộc)"
                    className="w-full min-h-16 rounded-md border border-input bg-background px-3 py-2 text-sm"
                  />
                  <div className="flex gap-2">
                    <Button type="submit" variant="destructive" size="sm" disabled={submittingReport}>
                      {submittingReport ? "Đang gửi..." : "Gửi báo cáo"}
                    </Button>
                    <Button type="button" variant="ghost" size="sm" onClick={() => setShowReport(false)}>
                      Hủy
                    </Button>
                  </div>
                </form>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

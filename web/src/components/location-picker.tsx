"use client";

import { useEffect, useState, useCallback } from "react";
import { MapPin, ChevronDown, Search, Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface Province {
  code: string;
  name: string;
}

interface Ward {
  code: string;
  name: string;
  provinceCode: string;
}

interface LocationPickerProps {
  initialProvince?: string;
  initialWard?: string;
  onChanged: (province: string | undefined, ward: string | undefined) => void;
}

function removeDiacritics(str: string): string {
  const withDiacritics =
    "àáảãạăắằẳẵặâấầẩẫậèéẻẽẹêếềểễệìíỉĩịòóỏõọôốồổỗộơớờởỡợùúủũụưứừửữựỳýỷỹỵđ";
  const withoutDiacritics =
    "aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd";
  let result = str;
  for (let i = 0; i < withDiacritics.length; i++) {
    result = result.replaceAll(withDiacritics[i], withoutDiacritics[i]);
  }
  return result;
}

export default function LocationPicker({
  initialProvince,
  initialWard,
  onChanged,
}: LocationPickerProps) {
  const [provinces, setProvinces] = useState<Province[]>([]);
  const [wards, setWards] = useState<Ward[]>([]);
  const [allWards, setAllWards] = useState<Ward[]>([]);

  const [selectedProvince, setSelectedProvince] = useState<Province | null>(null);
  const [selectedWard, setSelectedWard] = useState<Ward | null>(null);

  const [showProvinceModal, setShowProvinceModal] = useState(false);
  const [showWardModal, setShowWardModal] = useState(false);
  const [loading, setLoading] = useState(true);

  // Load CSV data
  useEffect(() => {
    async function loadCSV() {
      try {
        const res = await fetch("/vietnam_divisions.csv");
        const text = await res.text();
        const lines = text.split("\n").slice(3); // Skip 3 header rows

        const provinceMap = new Map<string, Province>();
        const wardList: Ward[] = [];

        for (const line of lines) {
          if (!line.trim()) continue;
          const cols = line.split(",");
          if (cols.length < 8) continue;

          const provinceCode = cols[2];
          const provinceName = cols[3];
          const wardCode = cols[6];
          const wardName = cols[7]?.trim();

          if (provinceCode && provinceName && !provinceMap.has(provinceCode)) {
            provinceMap.set(provinceCode, { code: provinceCode, name: provinceName });
          }

          if (wardCode && wardName) {
            wardList.push({ code: wardCode, name: wardName, provinceCode });
          }
        }

        const sortedProvinces = Array.from(provinceMap.values()).sort((a, b) =>
          a.name.localeCompare(b.name, "vi")
        );

        setProvinces(sortedProvinces);
        setAllWards(wardList);

        // Set initial values
        if (initialProvince) {
          const match = sortedProvinces.find((p) => p.name === initialProvince);
          if (match) {
            setSelectedProvince(match);
            const matchedWards = wardList.filter((w) => w.provinceCode === match.code);
            setWards(matchedWards);

            if (initialWard) {
              const wardMatch = matchedWards.find((w) => w.name === initialWard);
              if (wardMatch) setSelectedWard(wardMatch);
            }
          }
        }
      } catch {
        // ignore
      } finally {
        setLoading(false);
      }
    }
    loadCSV();
  }, [initialProvince, initialWard]);

  const handleProvinceSelect = useCallback(
    (province: Province) => {
      setSelectedProvince(province);
      setSelectedWard(null);
      setWards(allWards.filter((w) => w.provinceCode === province.code));
      setShowProvinceModal(false);
      onChanged(province.name, undefined);
    },
    [allWards, onChanged]
  );

  const handleWardSelect = useCallback(
    (ward: Ward) => {
      setSelectedWard(ward);
      setShowWardModal(false);
      onChanged(selectedProvince?.name, ward.name);
    },
    [selectedProvince, onChanged]
  );

  if (loading) {
    return <div className="h-11 rounded-md border border-input bg-muted animate-pulse" />;
  }

  return (
    <div className="grid grid-cols-2 gap-2">
      {/* Province selector */}
      <div className="relative">
        <button
          type="button"
          onClick={() => setShowProvinceModal(true)}
          className="flex items-center justify-between w-full h-11 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent transition-colors"
        >
          <span className={selectedProvince ? "text-foreground" : "text-muted-foreground"}>
            {selectedProvince?.name || "Tỉnh/Thành phố"}
          </span>
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        </button>
      </div>

      {/* Ward selector */}
      <div className="relative">
        <button
          type="button"
          onClick={() => selectedProvince && setShowWardModal(true)}
          disabled={!selectedProvince}
          className="flex items-center justify-between w-full h-11 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span className={selectedWard ? "text-foreground" : "text-muted-foreground"}>
            {selectedWard?.name || "Phường/Xã"}
          </span>
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        </button>
      </div>

      {/* Province search modal */}
      {showProvinceModal && (
        <SearchModal
          title="Chọn Tỉnh/Thành phố"
          items={provinces}
          getName={(p) => p.name}
          selected={selectedProvince}
          onSelect={handleProvinceSelect}
          onClose={() => setShowProvinceModal(false)}
        />
      )}

      {/* Ward search modal */}
      {showWardModal && (
        <SearchModal
          title="Chọn Phường/Xã"
          items={wards}
          getName={(w) => w.name}
          selected={selectedWard}
          onSelect={handleWardSelect}
          onClose={() => setShowWardModal(false)}
        />
      )}
    </div>
  );
}

function SearchModal<T>({
  title,
  items,
  getName,
  selected,
  onSelect,
  onClose,
}: {
  title: string;
  items: T[];
  getName: (item: T) => string;
  selected: T | null;
  onSelect: (item: T) => void;
  onClose: () => void;
}) {
  const [query, setQuery] = useState("");

  const filtered = query
    ? items.filter((item) =>
        removeDiacritics(getName(item).toLowerCase()).includes(
          removeDiacritics(query.toLowerCase())
        )
      )
    : items;

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50">
      <div className="bg-background w-full sm:max-w-md sm:rounded-lg rounded-t-lg max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h3 className="font-semibold">{title}</h3>
          <button type="button" onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Search */}
        <div className="p-3 border-b">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Tìm kiếm..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="pl-9 h-10"
              autoFocus
            />
          </div>
          <p className="text-xs text-muted-foreground mt-1">{filtered.length} kết quả</p>
        </div>

        {/* List */}
        <div className="overflow-y-auto flex-1">
          {filtered.map((item, i) => {
            const name = getName(item);
            const isSelected = selected !== null && getName(selected) === name;
            return (
              <button
                key={i}
                type="button"
                onClick={() => onSelect(item)}
                className={`w-full text-left px-4 py-3 text-sm hover:bg-accent transition-colors flex items-center justify-between ${
                  isSelected ? "bg-primary/5 text-primary" : ""
                }`}
              >
                <span>{name}</span>
                {isSelected && <Check className="h-4 w-4 text-primary" />}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

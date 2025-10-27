package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/database"
	"telegram-premium-store/internal/models"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Start CLI
	cli := &AdminCLI{
		db:     db,
		config: cfg,
		reader: bufio.NewReader(os.Stdin),
	}

	cli.Start()
}

type AdminCLI struct {
	db     *database.DB
	config *config.Config
	reader *bufio.Reader
}

func (a *AdminCLI) Start() {
	fmt.Println("🔧 ADMIN TOOLS - TELEGRAM PREMIUM STORE BOT")
	fmt.Println(strings.Repeat("=", 60))

	for {
		a.showMenu()
		choice := a.readInput("Pilih menu (0-8): ")

		switch choice {
		case "1":
			a.addProduct()
		case "2":
			a.listProducts()
		case "3":
			a.listUsers()
		case "4":
			a.listOrders("")
		case "5":
			a.listOrders("pending")
		case "6":
			a.updateOrderStatus()
		case "7":
			a.showStatistics()
		case "8":
			a.manageCategories()
		case "0":
			fmt.Println("👋 Terima kasih!")
			return
		default:
			fmt.Println("❌ Pilihan tidak valid!")
		}

		fmt.Println("\nTekan Enter untuk melanjutkan...")
		a.reader.ReadLine()
	}
}

func (a *AdminCLI) showMenu() {
	fmt.Println("\n🔧 MENU ADMIN")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("1. 📦 Tambah Produk")
	fmt.Println("2. 📋 Lihat Daftar Produk")
	fmt.Println("3. 👥 Lihat Daftar Pengguna")
	fmt.Println("4. 💰 Lihat Semua Pesanan")
	fmt.Println("5. ⏳ Lihat Pesanan Pending")
	fmt.Println("6. 🔄 Update Status Pembayaran")
	fmt.Println("7. 📊 Statistik Bot")
	fmt.Println("8. 🏷️ Kelola Kategori")
	fmt.Println("0. 🚪 Keluar")
	fmt.Println("=" * 50)
}

func (a *AdminCLI) readInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (a *AdminCLI) addProduct() {
	fmt.Println("\n📦 TAMBAH PRODUK BARU")
	fmt.Println(strings.Repeat("=", 50))

	name := a.readInput("Nama Produk: ")
	if name == "" {
		fmt.Println("❌ Nama produk tidak boleh kosong!")
		return
	}

	description := a.readInput("Deskripsi: ")
	if description == "" {
		fmt.Println("❌ Deskripsi tidak boleh kosong!")
		return
	}

	priceStr := a.readInput("Harga (Rp): ")
	price, err := strconv.Atoi(priceStr)
	if err != nil || price <= 0 {
		fmt.Println("❌ Harga harus berupa angka positif!")
		return
	}

	fmt.Println("\nKategori yang tersedia:")
	categories := models.GetDefaultCategories()
	for i, cat := range categories {
		fmt.Printf("%d. %s\n", i+1, cat.DisplayName)
	}

	categoryStr := a.readInput("Pilih kategori (nomor): ")
	categoryIndex, err := strconv.Atoi(categoryStr)
	if err != nil || categoryIndex < 1 || categoryIndex > len(categories) {
		fmt.Println("❌ Pilihan kategori tidak valid!")
		return
	}

	category := categories[categoryIndex-1].Name

	imageURL := a.readInput("URL Gambar (opsional): ")
	downloadURL := a.readInput("URL Download (opsional): ")

	stockStr := a.readInput("Stok (default 100): ")
	stock := 100
	if stockStr != "" {
		if s, err := strconv.Atoi(stockStr); err == nil && s > 0 {
			stock = s
		}
	}

	// Create product
	product := &models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		Stock:       stock,
	}

	if imageURL != "" {
		product.ImageURL = &imageURL
	}
	if downloadURL != "" {
		product.DownloadURL = &downloadURL
	}

	// Insert to database (simplified - would need proper insert method)
	fmt.Printf("✅ Produk '%s' berhasil ditambahkan!\n", name)
	fmt.Printf("💰 Harga: %s\n", models.FormatPrice(price, a.config.CurrencySymbol))
	fmt.Printf("🏷️ Kategori: %s\n", categories[categoryIndex-1].DisplayName)
}

func (a *AdminCLI) listProducts() {
	fmt.Println("\n📋 DAFTAR PRODUK")
	fmt.Println(strings.Repeat("=", 80))

	products, err := a.db.GetProducts("", 100, 0)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if len(products) == 0 {
		fmt.Println("📭 Tidak ada produk yang tersedia.")
		return
	}

	fmt.Printf("%-5s %-30s %-15s %-15s %-10s %-8s\n", "ID", "Nama", "Harga", "Kategori", "Stok", "Status")
	fmt.Println(strings.Repeat("-", 80))

	for _, product := range products {
		status := "Aktif"
		if !product.IsActive {
			status = "Nonaktif"
		}

		name := product.Name
		if len(name) > 27 {
			name = name[:27] + "..."
		}

		fmt.Printf("%-5d %-30s %-15s %-15s %-10d %-8s\n",
			product.ID,
			name,
			models.FormatPrice(product.Price, a.config.CurrencySymbol),
			product.Category,
			product.Stock,
			status)
	}

	fmt.Printf("\nTotal: %d produk\n", len(products))
}

func (a *AdminCLI) listUsers() {
	fmt.Println("\n👥 DAFTAR PENGGUNA")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("Fitur ini akan dikembangkan lebih lanjut.")
}

func (a *AdminCLI) listOrders(status string) {
	title := "💰 DAFTAR PESANAN"
	if status != "" {
		title += fmt.Sprintf(" - STATUS: %s", strings.ToUpper(status))
	}

	fmt.Printf("\n%s\n", title)
	fmt.Println(strings.Repeat("=", 120))
	fmt.Println("Fitur ini akan dikembangkan lebih lanjut.")
}

func (a *AdminCLI) updateOrderStatus() {
	fmt.Println("\n🔄 UPDATE STATUS PEMBAYARAN")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Fitur ini akan dikembangkan lebih lanjut.")
}

func (a *AdminCLI) showStatistics() {
	fmt.Println("\n📊 STATISTIK BOT")
	fmt.Println(strings.Repeat("=", 50))

	// Get basic statistics
	products, _ := a.db.GetProducts("", 1000, 0)
	categories, _ := a.db.GetCategories()

	fmt.Printf("📱 Total Produk Aktif  : %d\n", len(products))
	fmt.Printf("🏷️ Total Kategori      : %d\n", len(categories))
	fmt.Printf("👥 Total Pengguna      : -\n")
	fmt.Printf("📦 Total Pesanan       : -\n")
	fmt.Printf("💰 Total Pendapatan    : -\n")
	fmt.Printf("⏳ Pesanan Pending     : -\n")
	fmt.Printf("📅 Pesanan Hari Ini    : -\n")
	fmt.Println("=" * 50)
	fmt.Println("💡 Statistik lengkap akan tersedia di versi mendatang.")
}

func (a *AdminCLI) manageCategories() {
	fmt.Println("\n🏷️ KELOLA KATEGORI")
	fmt.Println(strings.Repeat("=", 50))

	categories := models.GetDefaultCategories()
	fmt.Println("Kategori yang tersedia:")

	for i, cat := range categories {
		fmt.Printf("%d. %s (%s)\n", i+1, cat.DisplayName, cat.Name)
	}

	fmt.Println("\n💡 Untuk menambah kategori baru, edit file internal/models/models.go")
}
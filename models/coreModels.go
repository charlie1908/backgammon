package models

// Zar atilinca tas hareket edebilir mi ?
// Kirik varsa FromBar: true, CanEnterFromBar: true Kirik giriyor mu, EnterableDice: []int Hangi zarlar ile kirik giriyor.
type MoveCheckResult struct {
	Allowed         bool  // Global Hareket edebilir sonucu
	FromBar         bool  // Kirik var ise ?
	CanEnterFromBar bool  // Kirik girebiliyor mu ?
	EnterableDice   []int //Kirik hangi zarlar ile giriyor ?
	CanMoveNormally bool  // Kirik yok pul oraya hareket edebiliyor mu ?
	NormalDice      []int // Zar mesafasi ile uyum kontrolu..
}

const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errorPickup: false,
    errorItems: false,
    errorCharity: false,
    hasError: false,
    locating: false,
    locateError: false,
    donate: {
      itemTypes: [],
      charityTypes: [],
      anyCharityType: null,
      proximity: '25',
      lat: null,
      long: null,
      pickupDropoff: null,
      zip: null,
    }
  },
  mounted() {
    try {
      donate = JSON.parse(localStorage.getItem('donate'));
      if (donate != null) {
        this.donate = donate
      }
    } catch(e) {
      localStorage.removeItem('donate')
    }
  },
  methods:{
    locateUser: function(e) {
      e.preventDefault();
      self = this

      if (navigator.geolocation) {
        this.locating = true
        var success =  function(pos) {
          self.locating = false
          console.log("got position" + pos)
        }
        var error = function(err) {
          console.log(err)
          self.locating = false
          self.locateError = true
          console.log("geo error " + err )
        }
        var options = {
          enableHighAccuracy: false,
          timeout: 2000,
          maximumAge: 1000*60*60
        }
        navigator.geolocation.getCurrentPosition(success, error, options)
      } else {
        console.log("Geolocation is not supported by this browser.")
      }
    },
    checkForm: function (e) {
      e.preventDefault();
      console.log(this.donate.anyCharityType)

      this.hasError = false
      this.errorPickup = false
      this.errorItems = false
      this.errorCharity = false

      if (!this.donate.pickupDropoff) {
        this.errorPickup = true
        this.hasError = true
      }

      if (this.donate.itemTypes.length == 0) {
        this.errorItems = true
        this.hasError = true
      }

      if (this.donate.charityTypes.length == 0 && !this.donate.anyCharityType) {
        this.errorCharity = true
        this.hasError = true
      }

      if (this.hasError) {
        return
      }

      this.saveStep();
      window.location.assign("/donate/results")
    },
    reset(e) {
      e.preventDefault()

      this.hasError = false
      this.errorPickup = false
      this.errorItems = false
      this.errorCharity = false

      this.donate.itemTypes = []
      this.donate.charityTypes = []
      this.donate.anyCharityType = null
      this.donate.pickupDropoff = null
      console.log('reset')
      localStorage.removeItem('donate')
    },
    saveStep() {
      localStorage.setItem('donate', JSON.stringify(this.donate))
    },
  }
})

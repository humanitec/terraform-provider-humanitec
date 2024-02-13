resource "humanitec_agent" "example" {
  id          = "agent-id"
  description = "Demo Agent"
  public_keys = [
    {
      key = "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtVChdg1SCUsMEs9Zuiuj\nRCNF7yGbLc+7GxchgsLrRhvqRjLkNh/757rS5Xb6fn5PSdV34iKsM1t/DNrohD6n\n/qO5CFT9PmJscG92ONNrma9Q+G2VqgOcQTBsvROnOXt3/sz3KxoVg7PH+dvpPOc2\na3vYI094OQ9290BtORer0gjdiacCadXlIucrfwQHHns5FUv8kui+AJ/EMRCGANkL\nW7V8sgrDJsazd9K+kZt6nBR2oalzmkTdS37hP1CqK2UzzZg+W2g6MjHIe6XfYqJa\nRnOSFM6sqRSXcgNH9ClUcCBvoDXtC1UxwIqstaHqXiEYiBntpHc9b7YxEfAs0Y3K\n574vbkN45hnx4q1je2Ajipfi6rCrD5krCZ3m00NtoddgjuTL4a5p9UxmC99WXacu\nxflmkpdYjI4fqvbZYBjc4JWuW95iW213BN2dlqQaEIKhepjROXc4D9AhMTjo4Vlr\nPrl32RuSLc+ZepP3KWxC/1X92PvHsrgQ09X4a1qZY8iBG468G6v2jOeJYTWDcXR+\nbEcw9ziXQtqwnI+PZEzDAJKNYA1VGSGQ1cr27aNUH6HcEb05mwIHXom7fYL1sEye\nWfJuEsiO2vlIh5L8FbyuNjafAPJHtw3TJAcq9JY7RGgM9fhc2NKLKnIKARBsycCS\nF+tcM6Fm/SqGiW7McuKcGsECAwEAAQ==\n-----END PUBLIC KEY-----\n"
    }
  ]
}
